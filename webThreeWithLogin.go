package balance

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

func webThreeWithLogin(c *http.Client, login, password string) (*Balance, error) {
	// replace +44 to 0
	if strings.HasPrefix(login, "+44") {
		login = "0" + login[3:]
	}

	// get cookies
	{
		req, err := http.NewRequest("GET", "https://www.three.co.uk/My3Account/My3/Login", nil)
		if err != nil {
			return nil, errors.Wrap(err, "account NewRequest")
		}

		resp, err := c.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "account Do")
		}
		_ = resp.Body.Close()
	}

	// get login page
	var authUrl string
	var lt string
	{
		req, err := http.NewRequest("GET", "https://sso.three.co.uk/mylogin/?service=https%3A%2F%2Fwww.three.co.uk%2FThreePortal%2Fappmanager%2FThree%2FSelfcareUk&resource=portlet", nil)
		if err != nil {
			return nil, errors.Wrap(err, "account NewRequest")
		}

		resp, err := c.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "account Do")
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "login get cookies NewDocumentFromReader")
		}

		val, ok := doc.Find(`form#securityCheck > .login > input[name="lt"]`).First().Attr("value")
		if !ok {
			err = errors.Errorf("could not extract lt value")

			// try get redirect link
			var ok bool
			authUrl, ok = doc.SetHtml(doc.Find("div.P00_id > noscript").Text()).Find("p > a").First().Attr("href")
			if !ok {
				return nil, errors.Wrap(err, "could not extract redirect url from response")
			}
		}

		lt = val
	}

	// post
	if authUrl == "" {
		data := make(url.Values)
		data.Set("username", login)
		data.Set("password", password)
		data.Set("lt", lt)
		resp, err := c.PostForm("https://sso.three.co.uk/mylogin/?service=https%3A%2F%2Fwww.three.co.uk%2FThreePortal%2Fappmanager%2FThree%2FSelfcareUk&resource=portlet", data)
		if err != nil {
			return nil, errors.Wrap(err, "login Get")
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "login post NewDocumentFromReader")
		}

		var ok bool
		authUrl, ok = doc.SetHtml(doc.Find("div.P00_id > noscript").Text()).Find("p > a").First().Attr("href")
		if !ok {
			return nil, errors.Errorf("could not extract redirect url from response")
		}
	}

	// other
	return webThreeBase(c, authUrl)
}
