package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type VivoxChannelType int32

const (
	Echo                             = 0
	LocalPositional VivoxChannelType = 1 // Local positional channel, positional, everyone can speak.
)

type VivoxChannelProperties struct {
	Id                     string           `json:"id,omitempty"`
	AudibleDistance        int              `json:"audibleDistance,omitempty"`
	ConversationalDistance int              `json:"conversationalDistance,omitempty"`
	AudioFadeIntensity     float32          `json:"audioFadeIntensity,omitempty"`
	AudioFadeModel         int              `json:"audioFadeModel,omitempty"`
	ChannelType            VivoxChannelType `json:"channelType,omitempty"`
}

type VivoxTokenPayload struct {
	Issuer            string                 `json:"iss,omitempty"`     // Issuer who provides the token.
	ExpiresAt         int64                  `json:"exp,omitempty"`     // Time of token expiration.
	Action            string                 `json:"vxa,omitempty"`     // Action this token is for.
	Id                int64                  `json:"vxi,omitempty"`     // Sequence id.
	Subject           string                 `json:"sub,omitempty"`     // User this token is for.
	From              string                 `json:"f,omitempty"`       // User this token is from.
	To                string                 `json:"t,omitempty"`       // Channel this token is for.
	Server            string                 `json:"server,omitempty"`  // Server to connect to.
	ChannelProperties VivoxChannelProperties `json:"channelProperties"` // Positional channel properties.
	Token             string                 `json:"token,omitempty"`   // Token issued for the user.
}

// region Vivox Credentials
const (
	vivoxIssuer      string = "xxxxxxxx"
	vivoxDomain      string = "xxxxxxxx"
	vivoxServer      string = "xxxxxxxx"
	vivoxSecret      string = "xxxxxxxx"
	vivoxAdminUserId string = "xxxxxxxx"
	vivoxAdminPasswd string = "xxxxxxxx"
)

//endregion

const (
	vivoxActionLogin  = "login"
	vivoxActionJoin   = "join"
	vivoxActionMute   = "mute"
	vivoxActionUnmute = "unmute"
	vivoxActionKick   = "kick"
)

const (
	vivoxLoginApi   = "/api2/viv_signin.php"
	vivoxChannelApi = "/api2/viv_chan_cmd.php"
)

const minAudibleDistance = 2700           // > 0
const minConversationalDistance = 90      // > 0
const minAudioFadeIntensity float32 = 0.1 // > 0

const defaultAudibleDistance = 8100           // > 2700
const defaultConversationalDistance = 270     // > 90
const defaultAudioFadeIntensity float32 = 0.5 // > 0
const defaultAudioFadeModel = 1               // 1, 2, 3

const defaultTokenExpirationTimespan = 60 * time.Second

// region Client Login Token
var loginTokenSerial int64 = 1

func GetLoginToken(payload VivoxTokenPayload) (string, error) {
	expiresAt := time.Now().Add(defaultTokenExpirationTimespan).Unix()

	payload.From = GetUserUri(payload.From)
	payload.Id = loginTokenSerial
	payload.ExpiresAt = expiresAt
	payload.Action = vivoxActionLogin
	payload.Issuer = vivoxIssuer
	payload.Server = vivoxServer

	token, err := GenerateVivoxToken(vivoxIssuer, expiresAt, vivoxActionLogin, loginTokenSerial, "", payload.From, "", vivoxSecret)
	if err != nil {
		return `{"status":"error"}`, err
	}

	payload.Token = token

	loginTokenSerial++

	jsonPayload, err := json.Marshal(payload)

	return string(jsonPayload), nil
}

//endregion

// region Client Join Token
var joinTokenSerial int64 = 1

func GetJoinToken(payload VivoxTokenPayload) (string, error) {
	expiresAt := time.Now().Add(defaultTokenExpirationTimespan).Unix()

	payload.From = GetUserUri(payload.From)
	payload.To = GetChannelUri(&payload.ChannelProperties)
	payload.Id = joinTokenSerial
	payload.ExpiresAt = expiresAt
	payload.Action = vivoxActionJoin
	payload.Issuer = vivoxIssuer
	payload.Server = vivoxServer

	token, err := GenerateVivoxToken(vivoxIssuer, expiresAt, vivoxActionJoin, joinTokenSerial, "", payload.From, payload.To, vivoxSecret)
	if err != nil {
		return `{"status":"error"}`, err
	}

	payload.Token = token

	joinTokenSerial++
	jsonPayload, err := json.Marshal(payload)
	return string(jsonPayload), nil
}

//endregion

// region Server-to-Server Data Structures
type vivoxXmlResponse struct {
	XMLName xml.Name               `xml:"response"`
	Root    vivoxLoginResponseRoot `xml:"level0"`
}

type vivoxLoginResponseRoot struct {
	XMLName xml.Name                 `xml:"level0"`
	Status  vivoxLoginResponseStatus `xml:"status"`
	Body    vivoxLoginResponseBody   `xml:"body"`
}

type vivoxLoginResponseStatus struct {
	XMLName xml.Name `xml:"status"`
	Status  string   `xml:"status"`
}

type vivoxLoginResponseBody struct {
	XMLName        xml.Name `xml:"body"`
	Code           int      `xml:"code"`
	Message        string   `xml:"msg"`
	Status         string   `xml:"status"`
	SipQVal        int      `xml:"sip_qval"`
	Login          string   `xml:"loginname"`
	DisplayName    string   `xml:"displayname"`
	AccountId      int      `xml:"account_id"`
	AuthToken      string   `xml:"auth_token"`
	SipAccount     string   `xml:"sip_acct"`
	SipUri         string   `xml:"sip_uri"`
	SipDisplayName string   `xml:"sip_displayname"`
}

//endregion

// region Server-to-Server Login
func RequestLogin() (string, error) {
	apiUrl := vivoxServer
	resource := vivoxLoginApi
	data := url.Values{}
	data.Set("userid", vivoxAdminUserId)
	data.Set("pwd", vivoxAdminPasswd)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return "error", err
	}

	if resp.StatusCode != 200 {
		return "error", fmt.Errorf("request error: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)

	var vxResponse vivoxXmlResponse
	err = xml.Unmarshal(responseBytes, &vxResponse)
	if err != nil {
		return "error", err
	}

	fmt.Printf(vxResponse.Root.Body.AuthToken)

	return vxResponse.Root.Body.AuthToken, err
}

//endregion

// region Server-to-Server Mute
var muteTokenSerial int64 = 1

func RequestMute(payload VivoxTokenPayload) (string, error) {
	authToken, err := RequestLogin()
	if err != nil {
		return `{"status":"error"}`, err
	}

	expiresAt := time.Now().Add(defaultTokenExpirationTimespan).Unix()

	payload.From = GetUserUri(payload.From)
	payload.To = GetChannelUri(&payload.ChannelProperties)
	payload.Subject = GetUserUri(payload.Subject)
	payload.Id = muteTokenSerial
	payload.ExpiresAt = expiresAt
	payload.Action = vivoxActionMute
	payload.Issuer = vivoxIssuer
	payload.Server = vivoxServer

	token, err := GenerateVivoxToken(vivoxIssuer, expiresAt, vivoxActionMute, muteTokenSerial, payload.Subject, payload.From, payload.To, vivoxSecret)
	if err != nil {
		return `{"status":"error"}`, err
	}

	muteTokenSerial++

	payload.Token = token

	apiUrl := vivoxServer
	resource := vivoxChannelApi
	data := url.Values{}
	data.Set("access_token", token)
	data.Set("auth_token", authToken)
	data.Set("chan_uri", payload.To)
	data.Set("user_uri", payload.Subject)
	data.Set("mode", "mute")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return `{"status":"error"}`, err
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return `{"status":"error"}`, fmt.Errorf("request error %d: %s", resp.StatusCode, responseBytes)
	}

	var vxResponse vivoxXmlResponse
	err = xml.Unmarshal(responseBytes, &vxResponse)
	if err != nil {
		return `{"status":"error"}`, err
	}

	if vxResponse.Root.Status.Status == "OK" {
		return `{"status":"ok"}`, nil
	} else {
		return fmt.Sprintf(`{"status":"error","code":"%d","message":"%s"}`, vxResponse.Root.Body.Code, vxResponse.Root.Body.Message), err
	}
}

var unmuteTokenSerial int64 = 1

func RequestUnmute(payload VivoxTokenPayload) (string, error) {
	authToken, err := RequestLogin()
	if err != nil {
		return `{"status":"error"}`, err
	}

	expiresAt := time.Now().Add(defaultTokenExpirationTimespan).Unix()

	payload.From = GetUserUri(payload.From)
	payload.To = GetChannelUri(&payload.ChannelProperties)
	payload.Subject = GetUserUri(payload.Subject)
	payload.Id = muteTokenSerial
	payload.ExpiresAt = expiresAt
	payload.Action = vivoxActionUnmute
	payload.Issuer = vivoxIssuer
	payload.Server = vivoxServer

	token, err := GenerateVivoxToken(vivoxIssuer, expiresAt, vivoxActionUnmute, unmuteTokenSerial, payload.Subject, payload.From, payload.To, vivoxSecret)
	if err != nil {
		return `{"status":"error"}`, err
	}

	unmuteTokenSerial++

	payload.Token = token

	apiUrl := vivoxServer
	resource := vivoxChannelApi
	data := url.Values{}
	data.Set("access_token", token)
	data.Set("auth_token", authToken)
	data.Set("chan_uri", payload.To)
	data.Set("user_uri", payload.Subject)
	data.Set("mode", "unmute")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return `{"status":"error"}`, err
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return `{"status":"error"}`, fmt.Errorf("request error %d: %s", resp.StatusCode, responseBytes)
	}

	var vxResponse vivoxXmlResponse
	err = xml.Unmarshal(responseBytes, &vxResponse)
	if err != nil {
		return `{"status":"error"}`, err
	}

	if vxResponse.Root.Status.Status == "OK" {
		return `{"status":"ok"}`, nil
	} else {
		return fmt.Sprintf(`{"status":"error","message":"%s"}`, vxResponse.Root.Body.Message), err
	}
}

//endregion

// region Server-to-Server Kick
var kickTokenSerial int64 = 1

func RequestKick(payload VivoxTokenPayload) (string, error) {
	authToken, err := RequestLogin()
	if err != nil {
		return `{"status":"error"}`, err
	}

	expiresAt := time.Now().Add(defaultTokenExpirationTimespan).Unix()

	payload.From = GetUserUri(payload.From)
	payload.To = GetChannelUri(&payload.ChannelProperties)
	payload.Subject = GetUserUri(payload.Subject)
	payload.Id = kickTokenSerial
	payload.ExpiresAt = expiresAt
	payload.Action = vivoxActionKick
	payload.Issuer = vivoxIssuer
	payload.Server = vivoxServer

	token, err := GenerateVivoxToken(vivoxIssuer, expiresAt, vivoxActionKick, kickTokenSerial, payload.Subject, payload.From, payload.To, vivoxSecret)
	if err != nil {
		return `{"status":"error"}`, err
	}

	kickTokenSerial++

	payload.Token = token

	apiUrl := vivoxServer
	resource := vivoxChannelApi
	data := url.Values{}
	data.Set("access_token", token)
	data.Set("auth_token", authToken)
	data.Set("chan_uri", payload.To)
	data.Set("user_uri", payload.Subject)
	data.Set("mode", "kick")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return `{"status":"error"}`, err
	}

	defer resp.Body.Close()
	responseBytes, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return `{"status":"error"}`, fmt.Errorf("request error %d: %s", resp.StatusCode, responseBytes)
	}

	var vxResponse vivoxXmlResponse
	err = xml.Unmarshal(responseBytes, &vxResponse)
	if err != nil {
		return `{"status":"error"}`, err
	}

	if vxResponse.Root.Status.Status == "OK" {
		return `{"status":"ok"}`, nil
	} else {
		return fmt.Sprintf(`{"status":"error","message":"%s"}`, vxResponse.Root.Body.Message), err
	}
}

//endregion

// region Get URI Helpers
func GetUserUri(userId string) string {
	return fmt.Sprintf("sip:.%s.%s.@%s", vivoxIssuer, userId, vivoxDomain)
}

func GetChannelUri(channelProperties *VivoxChannelProperties) string {
	var prefix string
	var properties string
	if channelProperties.ChannelType == Echo {
		prefix = "sip:confctl-e-"
		properties = ""
	} else if channelProperties.ChannelType == LocalPositional {
		prefix = "sip:confctl-d-"

		if channelProperties.AudibleDistance <= minAudibleDistance {
			channelProperties.AudibleDistance = defaultAudibleDistance
		}

		if channelProperties.ConversationalDistance <= minConversationalDistance {
			channelProperties.ConversationalDistance = defaultConversationalDistance
		}

		if channelProperties.AudioFadeIntensity <= minAudioFadeIntensity {
			channelProperties.AudioFadeIntensity = defaultAudioFadeIntensity
		}

		if !(channelProperties.AudioFadeModel >= 1 && channelProperties.AudioFadeModel <= 3) {
			channelProperties.AudioFadeModel = defaultAudioFadeModel
		}

		properties = strings.ReplaceAll(fmt.Sprintf("!p-%d-%d-%.3f-%d", channelProperties.AudibleDistance, channelProperties.ConversationalDistance, channelProperties.AudioFadeIntensity, channelProperties.AudioFadeModel), ",", ".")
	} else {
		prefix = "sip:confctl-g-"
		properties = ""
	}
	return fmt.Sprintf("%s%s.%s%s@%s", prefix, vivoxIssuer, channelProperties.Id, properties, vivoxDomain)
}

//endregion

// region Token Helper
func GenerateVivoxToken(issuer string, expiration int64, vxa string, serial int64, subject string, fromUri string, toUri string, key string) (string, error) {
	// Encode the empty header.
	encodedHeader := removePadding(base64.URLEncoding.EncodeToString([]byte("{}")))

	// Form the payload.
	var jsonPayload = "{"
	jsonPayload += fmt.Sprintf("\"iss\":\"%s\",", issuer)
	jsonPayload += fmt.Sprintf("\"exp\":%d,", expiration)
	jsonPayload += fmt.Sprintf("\"vxa\":\"%s\",", vxa)
	jsonPayload += fmt.Sprintf("\"vxi\":%d,", serial)
	if subject != "" {
		jsonPayload += fmt.Sprintf("\"vxi\":\"%s\",", subject)
	}
	if toUri != "" {
		jsonPayload += fmt.Sprintf("\"t\":\"%s\",", toUri)
	}
	if fromUri != "" {
		jsonPayload += fmt.Sprintf("\"f\":\"%s\",", fromUri)
	}
	jsonPayload = strings.TrimRight(jsonPayload, ",")
	jsonPayload += "}"
	jsonPayload = removeSpaces(jsonPayload)

	// Encode the payload.
	encodedPayload := removePadding(base64.URLEncoding.EncodeToString([]byte(jsonPayload)))
	toSign := encodedHeader + "." + encodedPayload

	// Compute the signature.
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write([]byte(toSign))
	if err != nil {
		return "", err
	}

	signature := removePadding(base64.URLEncoding.EncodeToString(h.Sum(nil)))

	return toSign + "." + signature, nil
}

//endregion

// region String Helpers
func removePadding(str string) string {
	return strings.ReplaceAll(str, "=", "")
}

func removeSpaces(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

//endregion
