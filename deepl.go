package deepl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Deepl struct {
	baseURL string
	apiKey  string
}

type TranslationPayload struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang"`
	GlossaryID string   `json:"glossary_id"`
}

type TranslationResponse struct {
	Translations []Translation `json:"translations"`
}

type Translation struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}

func New(baseURL, apiKey string) (*Deepl, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	} else if strings.HasSuffix(baseURL, "/") {
		return nil, fmt.Errorf("baseURL must not end with a slash")
	}

	return &Deepl{apiKey: apiKey, baseURL: baseURL}, nil
}

func (d *Deepl) Translate(text []string, sourceLang, targetLang string, glossaryID string) ([]Translation, error) {

	payload := TranslationPayload{
		Text:       text,
		TargetLang: targetLang,
		SourceLang: sourceLang,
		GlossaryID: glossaryID,
	}

	j, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/translate", d.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", d.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData := TranslationResponse{}

	err = json.NewDecoder(resp.Body).Decode(&responseData)

	if err != nil {
		return nil, err
	}

	return responseData.Translations, nil
}

func (d *Deepl) GetGlossaries() (string, error) {

	url := fmt.Sprintf("%s/glossaries", d.baseURL)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", d.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	return buf.String(), nil
}

type CreateGlossaryPayload struct {
	Name          string `json:"name"`
	SourceLang    string `json:"source_lang"`
	TargetLang    string `json:"target_lang"`
	EntriesFormat string `json:"entries_format"`
	Entries       string `json:"entries"`
}

type Glossary struct {
	ID           string `json:"glossary_id"`
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	SourceLang   string `json:"source_lang"`
	TargetLang   string `json:"target_lang"`
	CreationTime string `json:"creation_time"`
	EntryCount   int    `json:"entry_count"`
}

func (d *Deepl) CreateGlossary(name, sourceLang, targetLang string, entriesTSV io.Reader) (*Glossary, error) {

	buf := new(strings.Builder)
	_, err := io.Copy(buf, entriesTSV)

	if err != nil {
		return nil, err
	}

	entries := strings.ReplaceAll(buf.String(), "\r\n", "\n")

	payload := CreateGlossaryPayload{
		Name:          name,
		SourceLang:    sourceLang,
		TargetLang:    targetLang,
		EntriesFormat: "tsv",
		Entries:       entries,
	}

	j, err := json.Marshal(payload)

	fmt.Println(string(j))

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/glossaries", d.baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(j))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", d.apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	log.Printf("Response: %s", resp.Status)

	if err != nil {
		return nil, err
	}

	createdGlossary := Glossary{}

	err = json.NewDecoder(resp.Body).Decode(&createdGlossary)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	return &createdGlossary, nil
}
