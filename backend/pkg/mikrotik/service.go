package mikrotik

import (
	"fmt"
	"net"
	"time"

	"github.com/go-routeros/routeros"
)

// CheckIPInSecrets checks if an IP is already assigned in the router's PPP secrets
func CheckIPInSecrets(client *routeros.Client, ipAddress string) (string, error) {
	reply, err := client.Run("/ppp/secret/print", "?remote-address="+ipAddress)
	if err != nil {
		return "", err
	}
	if len(reply.Re) > 0 {
		return reply.Re[0].Map["name"], nil
	}
	return "", nil
}

// GetAllPPPSecrets fetches all PPP secrets from the router
func GetAllPPPSecrets(client *routeros.Client) ([]map[string]string, error) {
	reply, err := client.Run("/ppp/secret/print")
	if err != nil {
		return nil, err
	}
	var secrets []map[string]string
	for _, sentence := range reply.Re {
		secrets = append(secrets, sentence.Map)
	}
	return secrets, nil
}

// GetAllPPPProfiles fetches all PPP profiles from the router
func GetAllPPPProfiles(client *routeros.Client) ([]string, error) {
	reply, err := client.Run("/ppp/profile/print")
	if err != nil {
		return nil, err
	}
	var profiles []string
	for _, sentence := range reply.Re {
		if name, ok := sentence.Map["name"]; ok {
			profiles = append(profiles, name)
		}
	}
	return profiles, nil
}

// CreatePPPoESecret creates a new PPP secret in the router
func CreatePPPoESecret(client *routeros.Client, name, password, profile, ipAddress string) error {
	args := []string{
		"/ppp/secret/add",
		"=name=" + name,
		"=password=" + password,
		"=profile=" + profile,
		"=service=pppoe",
	}
	if ipAddress != "" {
		args = append(args, "=remote-address="+ipAddress)
	}
	_, err := client.RunArgs(args)
	return err
}

// UpdatePPPoESecret updates a PPP secret in the router. Finds the secret by oldName first.
func UpdatePPPoESecret(client *routeros.Client, oldName, newName, password, profile, ipAddress, disabled string) error {
	reply, err := client.Run("/ppp/secret/print", "?name="+oldName)
	if err != nil {
		return err
	}
	if len(reply.Re) == 0 {
		// Fallback to create if not found
		return CreatePPPoESecret(client, newName, password, profile, ipAddress)
	}

	secretID := reply.Re[0].Map[".id"]
	args := []string{
		"/ppp/secret/set",
		"=.id=" + secretID,
		"=name=" + newName,
		"=password=" + password,
		"=profile=" + profile,
		"=disabled=" + disabled,
	}
	if ipAddress != "" {
		args = append(args, "=remote-address="+ipAddress)
	} else {
		args = append(args, "=remote-address=")
	}
	_, err = client.RunArgs(args)
	return err
}

// DeletePPPoESecret deletes a PPP secret in the router by name
func DeletePPPoESecret(client *routeros.Client, name string) (bool, error) {
	reply, err := client.Run("/ppp/secret/print", "?name="+name)
	if err != nil {
		return false, err
	}
	if len(reply.Re) == 0 {
		return false, nil
	}
	secretID := reply.Re[0].Map[".id"]
	_, err = client.Run("/ppp/secret/remove", "=.id="+secretID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// RemoveActiveConnection disconnects the active PPP session by name
func RemoveActiveConnection(client *routeros.Client, name string) error {
	reply, err := client.Run("/ppp/active/print", "?name="+name)
	if err != nil {
		return err
	}
	if len(reply.Re) == 0 {
		return nil
	}
	activeID := reply.Re[0].Map[".id"]
	_, err = client.Run("/ppp/active/remove", "=.id="+activeID)
	return err
}

// TestConnection tests connection connectivity and retrieves basic information
func TestConnection(host string, port int, username, password string, timeout time.Duration) (string, string, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))
	client, err := routeros.NewClient(conn)
	if err != nil {
		return "", "", err
	}
	defer client.Close()

	err = client.Login(username, password)
	if err != nil {
		return "", "", err
	}

	// Fetch system identity
	replyId, err := client.Run("/system/identity/print")
	if err != nil {
		return "", "", err
	}
	identity := "Unknown"
	if len(replyId.Re) > 0 {
		identity = replyId.Re[0].Map["name"]
	}

	// Fetch system resource version
	replyRes, err := client.Run("/system/resource/print")
	if err != nil {
		return "", "", err
	}
	rosVersion := "N/A"
	if len(replyRes.Re) > 0 {
		rosVersion = replyRes.Re[0].Map["version"]
	}

	return identity, rosVersion, nil
}
