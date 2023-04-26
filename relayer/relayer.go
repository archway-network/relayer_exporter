package relayer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	ChainID   string
	Path      string
	ExpiresAt time.Time
}

func parsePaths(out io.Reader) ([]string, error) {
	paths := []string{}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, "->")
		if len(parts) < 2 {
			return nil, fmt.Errorf("parse error for line: %s", line)
		}

		parts = strings.Split(parts[0], " ")
		if len(parts) < 2 {
			return nil, fmt.Errorf("parse error for line: %s", line)
		}
		paths = append(paths, strings.TrimSpace(parts[1]))
	}

	return paths, nil
}

func parseClientsForPath(path string, out io.Reader) ([]Client, error) {
	clients := []Client{}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		r := regexp.MustCompile(`\((.+?)\)`)
		res := r.FindAllStringSubmatch(line, -1)

		err := fmt.Errorf("parse error for line: %s", line)

		if len(res) != 2 {
			return nil, err
		}
		if len(res[0]) != 2 {
			return nil, err
		}
		if len(res[1]) != 2 {
			return nil, err
		}

		expiresAt, err := time.Parse(time.RFC822, res[1][1])
		if err != nil {
			return nil, err
		}

		clients = append(clients, Client{ChainID: res[0][1], Path: path, ExpiresAt: expiresAt})
	}

	return clients, nil
}

func GetClients(relayerCmd string) ([]Client, error) {
	clients := []Client{}
	out, err := exec.Command(relayerCmd, []string{"paths", "list"}...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			log.Println(string(err.Stderr))
		}
		return nil, err
	}

	paths, err := parsePaths(bytes.NewBuffer(out))
	if err != nil {
		return nil, err
	}

	for _, p := range paths {
		out, err := exec.Command(relayerCmd, []string{"query", "clients-expiration", p}...).Output()
		if err != nil {
			if err, ok := err.(*exec.ExitError); ok {
				log.Println(string(err.Stderr))
			}
			continue
		}

		c, err := parseClientsForPath(p, bytes.NewBuffer(out))
		if err != nil {
			log.Println(err)
			continue
		}
		clients = append(clients, c...)
	}

	return clients, nil
}
