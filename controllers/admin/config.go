package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/owncast/owncast/controllers"
	"github.com/owncast/owncast/core"
	"github.com/owncast/owncast/core/data"
	"github.com/owncast/owncast/models"
	"github.com/owncast/owncast/utils"
	log "github.com/sirupsen/logrus"
)

// ConfigValue is a container object that holds a value, is encoded, and saved to the database.
type ConfigValue struct {
	Value interface{} `json:"value"`
}

// SetTags will handle the web config request to set tags.
func SetTags(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValues, success := getValuesFromRequest(w, r)
	if !success {
		return
	}

	var tagStrings []string
	for _, tag := range configValues {
		tagStrings = append(tagStrings, tag.Value.(string))
	}

	if err := data.SetServerMetadataTags(tagStrings); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetStreamTitle will handle the web config request to set the current stream title.
func SetStreamTitle(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	value := configValue.Value.(string)

	if err := data.SetStreamTitle(value); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}
	if value != "" {
		sendSystemChatAction(fmt.Sprintf("Stream title changed to **%s**", value), true)
	}
	controllers.WriteSimpleResponse(w, true, "changed")
}

func sendSystemChatAction(messageText string, ephemeral bool) {
	message := models.ChatEvent{}
	message.Body = messageText
	message.MessageType = models.ChatActionSent
	message.ClientID = "internal-server"
	message.Ephemeral = ephemeral
	message.SetDefaults()

	message.RenderBody()

	if err := core.SendMessageToChat(message); err != nil {
		log.Errorln(err)
	}
}

// SetServerName will handle the web config request to set the server's name.
func SetServerName(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetServerName(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetServerSummary will handle the web config request to set the about/summary text.
func SetServerSummary(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetServerSummary(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetServerWelcomeMessage will handle the web config request to set the welcome message text.
func SetServerWelcomeMessage(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetServerWelcomeMessage(strings.TrimSpace(configValue.Value.(string))); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetExtraPageContent will handle the web config request to set the page markdown content.
func SetExtraPageContent(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetExtraPageBodyContent(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetStreamKey will handle the web config request to set the server stream key.
func SetStreamKey(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetStreamKey(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetLogo will handle a new logo image file being uploaded.
func SetLogo(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	s := strings.SplitN(configValue.Value.(string), ",", 2)
	if len(s) < 2 {
		controllers.WriteSimpleResponse(w, false, "Error splitting base64 image data.")
		return
	}
	bytes, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	splitHeader := strings.Split(s[0], ":")
	if len(splitHeader) < 2 {
		controllers.WriteSimpleResponse(w, false, "Error splitting base64 image header.")
		return
	}
	contentType := strings.Split(splitHeader[1], ";")[0]
	extension := ""
	if contentType == "image/svg+xml" {
		extension = ".svg"
	} else if contentType == "image/gif" {
		extension = ".gif"
	} else if contentType == "image/png" {
		extension = ".png"
	} else if contentType == "image/jpeg" {
		extension = ".jpeg"
	}

	if extension == "" {
		controllers.WriteSimpleResponse(w, false, "Missing or invalid contentType in base64 image.")
		return
	}

	imgPath := filepath.Join("data", "logo"+extension)
	if err := ioutil.WriteFile(imgPath, bytes, 0644); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	if err := data.SetLogoPath("logo" + extension); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetNSFW will handle the web config request to set the NSFW flag.
func SetNSFW(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetNSFW(configValue.Value.(bool)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetFfmpegPath will handle the web config request to validate and set an updated copy of ffmpg.
func SetFfmpegPath(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	path := configValue.Value.(string)
	if err := utils.VerifyFFMpegPath(path); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	if err := data.SetFfmpegPath(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "changed")
}

// SetWebServerPort will handle the web config request to set the server's HTTP port.
func SetWebServerPort(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetHTTPPortNumber(configValue.Value.(float64)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "http port set")
}

// SetRTMPServerPort will handle the web config request to set the inbound RTMP port.
func SetRTMPServerPort(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetRTMPPortNumber(configValue.Value.(float64)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "rtmp port set")
}

// SetServerURL will handle the web config request to set the full server URL.
func SetServerURL(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetServerURL(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "server url set")
}

// SetDirectoryEnabled will handle the web config request to enable or disable directory registration.
func SetDirectoryEnabled(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetDirectoryEnabled(configValue.Value.(bool)); err != nil {
		controllers.WriteSimpleResponse(w, false, err.Error())
		return
	}
	controllers.WriteSimpleResponse(w, true, "directory state changed")
}

// SetStreamLatencyLevel will handle the web config request to set the stream latency level.
func SetStreamLatencyLevel(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		return
	}

	if err := data.SetStreamLatencyLevel(configValue.Value.(float64)); err != nil {
		controllers.WriteSimpleResponse(w, false, "error setting stream latency "+err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "set stream latency")
}

// SetS3Configuration will handle the web config request to set the storage configuration.
func SetS3Configuration(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	type s3ConfigurationRequest struct {
		Value models.S3 `json:"value"`
	}

	decoder := json.NewDecoder(r.Body)
	var newS3Config s3ConfigurationRequest
	if err := decoder.Decode(&newS3Config); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update s3 config with provided values")
		return
	}

	if newS3Config.Value.Enabled {
		if newS3Config.Value.Endpoint == "" || !utils.IsValidUrl((newS3Config.Value.Endpoint)) {
			controllers.WriteSimpleResponse(w, false, "s3 support requires an endpoint")
			return

		}

		if newS3Config.Value.AccessKey == "" || newS3Config.Value.Secret == "" {
			controllers.WriteSimpleResponse(w, false, "s3 support requires an access key and secret")
			return
		}

		if newS3Config.Value.Region == "" {
			controllers.WriteSimpleResponse(w, false, "s3 support requires a region and endpoint")
			return
		}

		if newS3Config.Value.Bucket == "" {
			controllers.WriteSimpleResponse(w, false, "s3 support requires a bucket created for storing public video segments")
			return
		}
	}

	data.SetS3Config(newS3Config.Value)
	controllers.WriteSimpleResponse(w, true, "storage configuration changed")

}

// SetStreamOutputVariants will handle the web config request to set the video output stream variants.
func SetStreamOutputVariants(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	type streamOutputVariantRequest struct {
		Value []models.StreamOutputVariant `json:"value"`
	}

	decoder := json.NewDecoder(r.Body)
	var videoVariants streamOutputVariantRequest
	if err := decoder.Decode(&videoVariants); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update video config with provided values "+err.Error())
		return
	}

	if err := data.SetStreamOutputVariants(videoVariants.Value); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update video config with provided values "+err.Error())
		return
	}

	controllers.WriteSimpleResponse(w, true, "stream output variants updated")
}

// SetSocialHandles will handle the web config request to set the external social profile links.
func SetSocialHandles(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	type socialHandlesRequest struct {
		Value []models.SocialHandle `json:"value"`
	}

	decoder := json.NewDecoder(r.Body)
	var socialHandles socialHandlesRequest
	if err := decoder.Decode(&socialHandles); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update social handles with provided values")
		return
	}

	if err := data.SetSocialHandles(socialHandles.Value); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update social handles with provided values")
		return
	}

	controllers.WriteSimpleResponse(w, true, "social handles updated")
}

// SetChatDisabled will disable chat functionality.
func SetChatDisabled(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		controllers.WriteSimpleResponse(w, false, "unable to update chat disabled")
		return
	}

	data.SetChatDisabled(configValue.Value.(bool))

	controllers.WriteSimpleResponse(w, true, "chat disabled status updated")
}

// SetVideoCodec will change the codec used for video encoding.
func SetVideoCodec(w http.ResponseWriter, r *http.Request) {
	if !requirePOST(w, r) {
		return
	}

	configValue, success := getValueFromRequest(w, r)
	if !success {
		controllers.WriteSimpleResponse(w, false, "unable to change video codec")
		return
	}

	if err := data.SetVideoCodec(configValue.Value.(string)); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update codec")
		return
	}

	controllers.WriteSimpleResponse(w, true, "video codec updated")
}

// SetExternalActions will set the 3rd party actions for the web interface.
func SetExternalActions(w http.ResponseWriter, r *http.Request) {
	type externalActionsRequest struct {
		Value []models.ExternalAction `json:"value"`
	}

	decoder := json.NewDecoder(r.Body)
	var actions externalActionsRequest
	if err := decoder.Decode(&actions); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update external actions with provided values")
		return
	}

	if err := data.SetExternalActions(actions.Value); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to update external actions with provided values")
	}

	controllers.WriteSimpleResponse(w, true, "external actions update")
}

// SetCustomStyles will set the CSS string we insert into the page.
func SetCustomStyles(w http.ResponseWriter, r *http.Request) {
	customStyles, success := getValueFromRequest(w, r)
	if !success {
		controllers.WriteSimpleResponse(w, false, "unable to update custom styles")
		return
	}

	data.SetCustomStyles(customStyles.Value.(string))

	controllers.WriteSimpleResponse(w, true, "custom styles updated")
}

// SetUsernameBlocklist will set the list of usernames we do not allow to use.
func SetUsernameBlocklist(w http.ResponseWriter, r *http.Request) {
	usernames, success := getValueFromRequest(w, r)
	if !success {
		controllers.WriteSimpleResponse(w, false, "unable to update custom styles")
		return
	}

	data.SetUsernameBlocklist(usernames.Value.(string))

	controllers.WriteSimpleResponse(w, true, "blocklist updated")
}

func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != controllers.POST {
		controllers.WriteSimpleResponse(w, false, r.Method+" not supported")
		return false
	}

	return true
}

func getValueFromRequest(w http.ResponseWriter, r *http.Request) (ConfigValue, bool) {
	decoder := json.NewDecoder(r.Body)
	var configValue ConfigValue
	if err := decoder.Decode(&configValue); err != nil {
		log.Warnln(err)
		controllers.WriteSimpleResponse(w, false, "unable to parse new value")
		return configValue, false
	}

	return configValue, true
}

func getValuesFromRequest(w http.ResponseWriter, r *http.Request) ([]ConfigValue, bool) {
	var values []ConfigValue

	decoder := json.NewDecoder(r.Body)
	var configValue ConfigValue
	if err := decoder.Decode(&configValue); err != nil {
		controllers.WriteSimpleResponse(w, false, "unable to parse array of values")
		return values, false
	}

	object := reflect.ValueOf(configValue.Value)

	for i := 0; i < object.Len(); i++ {
		values = append(values, ConfigValue{Value: object.Index(i).Interface()})
	}

	return values, true
}
