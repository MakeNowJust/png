package png

import (
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

func Parse(rawurl string) (Pinger, error) {
	if rawurl == "" {
		return nil, errors.New("cannot create pinger from empty string")
	}

	u, err := parseURL(rawurl)
	if err != nil {
		return nil, err
	}

	// When hostname is not specified, sets `127.0.0.1`.
	if u.Hostname() == "" {
		u.Host = "127.0.0.1:" + u.Port()
	}

	switch u.Scheme {
	case "http":
		fallthrough
	case "https":
		return &HTTPPinger{url: u}, nil

	case "redis":
		return parseRedis(u)

	default:
		return nil, errors.Errorf("invalid scheme: %s", u.Scheme)
	}
}

func parseURL(rawurl string) (u *url.URL, err error) {
	u, err = url.Parse(rawurl)
	if err != nil {
		err = errors.Wrapf(err, "failed in parsing URL: %#v", rawurl)
	} else {
		// - `u.Opaque != ""` is for `localhost:8080` case.
		// - `u.Scheme != ""` is for `localhost` case.
		if u.Opaque != "" || u.Scheme == "" {
			rawurl = "http://" + rawurl

			u, err = url.Parse(rawurl)
			err = errors.Wrapf(err, "failed in parsing URL: %#v", rawurl)
		}
	}

	return
}

func parseRedis(u *url.URL) (Pinger, error) {
	var password string
	port := 6379 // Redis well-known port
	var db int

	if user := u.User; user != nil {
		password, _ = user.Password()
		// TODO: should we treat username as password? It maybe useful but it maybe break consistent.
	}

	if portString := u.Port(); portString != "" {
		var err error
		port, err = strconv.Atoi(portString)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid port number: %s", portString)
		}
	}

	if path := u.Path; len(path) >= 2 {
		path = path[1:] // skip `/`

		var err error
		db, err = strconv.Atoi(path)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid db number: %#v", path)
		}
	}

	return &RedisPinger{
		hostname: u.Hostname(),
		port:     port,
		password: password,
		db:       db,
	}, nil
}
