#!/usr/bin/env bash

sudo apt update
sudo apt upgrade -y

sudo apt install wireguard -y

wg genkey | tee /etc/wireguard/server_private_key | wg pubkey >/etc/wireguard/server_public_key
wg genkey | tee /etc/wireguard/client_private_key | wg pubkey >/etc/wireguard/client_public_key

cat >/etc/wireguard/wg0.conf <<EOF
[Interface]
PrivateKey = $(cat /etc/wireguard/server_private_key)
ListenPort = 51820
Address = 10.0.0.1/24
PostUp = iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT; iptables -t nat -A POSTROUTING -o ens3 -j MASQUERADE
PostDown = iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT; iptables -t nat -D POSTROUTING -o ens3 -j MASQUERADE

[Peer]
PublicKey = $(cat /etc/wireguard/client_public_key)
AllowedIPs = 10.0.0.2/32

EOF

sudo chmod 700 /etc/wireguard/server_private_key

sudo nano /etc/sysctl.conf
sudo sysctl -p

sudo ufw allow ssh
sudo ufw allow 51820/udp
sudo ufw enable

sudo systemctl enable wg-quick@wg0.service
sudo systemctl start wg-quick@wg0.service

echo "
[Interface]
PrivateKey = $(cat /etc/wireguard/client_private_key)
Address = 10.0.0.2/32
DNS = 8.8.8.8

[Peer]
PublicKey = $(cat /etc/wireguard/server_public_key)
AllowedIPs = 0.0.0.0/0
Endpoint = VALUE:51820
PersistentKeepalive = 20

"
