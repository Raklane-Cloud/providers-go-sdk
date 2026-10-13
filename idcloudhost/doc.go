// Package idcloudhost is a client for the IDCloudHost cloud API
// (https://api.idcloudhost.com): virtual machines, block storage,
// private networks, floating IPs and firewalls.
//
// The API only answers requests from Indonesian IP addresses; call it
// from a host in Indonesia.
//
// Most resources live in one location (data centre). Client.Location
// selects it; empty uses the account's default location.
package idcloudhost
