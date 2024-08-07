package ute

import (
	"errors"
	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func uteLogin(client *http.Client, username, password string) error {

	log.Debug().Msg("Fetch login page from Identity Server...")

	res, err := client.Get("https://www.ute.com.uy/oauth")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	form, _, err := getFormInputsAndAction(string(data))
	if err != nil {
		return err
	}

	form.Set("Username", username)
	form.Set("Password", password)
	form.Add("RememberLogin", "false")
	form.Set("button", "login")

	log.Debug().Msg("Submit login credentials to Identity Server...")

	res, err = client.PostForm(res.Request.URL.String(), form)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := checkForLoginError(data); err != nil {
		return err
	}

	credentialsForm, postUrl, err := getFormInputsAndAction(string(data))
	if err != nil {
		return err
	}

	if credentialsForm.Get("id_token") == "" {
		return errors.New("login failed: id_token not found")
	}

	log.Debug().Msg("Identity Server login successful! Creating session on UTE website...")

	res, err = client.PostForm(postUrl, credentialsForm)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New("login failed: status code is not 200")
	}

	defer res.Body.Close()

	log.Debug().Msg("Login completed on UTE website!")

	return nil
}

func autoServicioLogin(client *http.Client) error {

	log.Debug().Msg("Fetch login page from AutoServicio...")

	res, err := client.Get("https://autoservicio.ute.com.uy/SelfService/SSvcController/loginidp")
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New("login failed: status code is not 200")
	}

	rawQuery := res.Request.URL.Fragment
	baseUrl, _ := url.Parse("https://autoservicio.ute.com.uy/SelfService/SSvcController/authenticate")
	baseUrl.RawQuery = rawQuery

	log.Debug().Msg("Submit login credentials to AutoServicio...")

	res, err = client.Get(baseUrl.String())
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.New("login failed: status code is not 200")
	}

	log.Debug().Msg("AutoServicio login successful!")

	return nil
}

func checkForLoginError(htmlData []byte) error {
	doc, err := htmlquery.Parse(strings.NewReader(string(htmlData)))
	if err != nil {
		return err
	}

	errorNode := htmlquery.FindOne(doc, `//div[contains(@class, "alert-danger")]/ul/li/text()`)
	if errorNode != nil {
		return errors.New(errorNode.Data)
	}

	return nil
}

func getFormInputsAndAction(html string) (url.Values, string, error) {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		return nil, "", err
	}

	formParams := url.Values{}
	inputNodes := htmlquery.Find(doc, "//input")
	for _, n := range inputNodes {
		formParams.Add(htmlquery.SelectAttr(n, "name"), htmlquery.SelectAttr(n, "value"))
	}

	form := htmlquery.FindOne(doc, "//form")
	action := htmlquery.SelectAttr(form, "action")

	return formParams, action, nil
}
