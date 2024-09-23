# local_dns

Local DNS for development use. Add and delete domain names via web API.  
`/etc/hosts` and `dnsmasq` require `sudo`, but this tool can be used without it.

## How to Build
```sh
go build
```

## Simple Usage

Start with the following command

```sh
./local_dns
```
Accepts DNS on port 2053 and web API on port 2080.
You can change the ports using the `--dns-port` and `--http-port` parameters.


## Example of Adding a Domain Name

To set `test.example.com` to localhost, use the following web API:


```sh
curl -POST http://localhost:2080/api --json '{"test.example.com":"127.0.0.1"}'
```

Verification
```sh
dig @localhost -p 2053 test.example.com
#:
#test.example.com. 60 IN A 127.0.0.1
#:
```

## Command

### Starting the Service

You can change the behavior of `local_dns` by specifying parameters when starting it.

```sh
# Open DNS on port 53 and HTTP on port 80
./local_dns --dns-port 53 --http-port 80

# Similarly, you can set it via environment variables
LOCALDNS_DNS_PORT=53 LOCALDNS_HTTP_PORT=80 ./local_dns

# Set a fallback DNS
./local_dns --fallback-ip 8.8.8.8
```

| Parameter         | Environment Variable      | Default | Description                                |
|:------------------|:--------------------------|:--------|:-------------------------------------------|
| `--dns-port`      | `LOCALDNS_DNS_PORT`       | 2053    | DNS port                                   |
| `--http-port`     | `LOCALDNS_HTTP_PORT`      | 2080    | HTTP port                                  |
| `--localhost-only`| `LOCALDNS_LOCALHOST_ONLY` | true    | Prohibit WebAPI access from non-127.0.0.1  |
| `--fallback-ip`   | `LOCALDNS_FALLBACK_IP`    | n/a     | Server to query if the domain name is not found |
| `--DEBUG`         | `LOCALDNS_DEBUG`          | false   | Output debug logs                          |

### webapi

You can add, delete, and check domain names using the web API.

```sh
# Get a list of domains held by local_dns
curl http://localhost:2080/api

# Set a new domain
curl -POST http://localhost:2080/api -d '{"hoge":"192.168.1.2"}'

# Delete a domain. Either the domain name or the IP address can be used
curl -DELETE http://localhost:2080/api?domain=hoge
curl -DELETE http://localhost:2080/api?address=192.168.1.2

# Delete all domains
curl -DELETE http://localhost:2080/api
```



## Example of Starting Nginx on Minikube

As a simple usage example, here is how to set up an Nginx server on Minikube and access it via `www.example.com` from a browser.

```
# minikube initialize
minikube start
minikube addons enable ingress

# setup nginx
kubectl create deployment nginx --image=nginx
kubectl expose deployment nginx --port=80 --target-port=80

# setup ingress
cat <<EOF > ingress-nginx.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: www.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx
            port:
              number: 80
EOF
kubectl apply -f ingress-nginx.yaml

# register minikube IP as www.example.com 
curl -POST http://localhost:2080/api --json "{\"www.example.com.\":\"$(minikube ip)\"}"
```

## Integration into Development Machine

Install in the execution directory of the machine as follows

```sh
go build
sudo mv local_dns /usr/local/bin/
sudo chown nobody:nogroup /usr/local/bin/local_dns
```



### Configuration for Ubuntu 22.04, 24.04, etc. (systemd-resolved)

Register and Start with systemctl

Create `/etc/systemd/system/local_dns.service` as follows
```
[Unit]
Description=local_dns
After=network.target

[Service]
ExecStart=/usr/local/bin/local_dns
Restart=always
User=nobody
Group=nogroup
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=GO_ENV=production
WorkingDirectory=/usr/local/bin

[Install]
WantedBy=multi-user.target
```

Start the local_dns.service
```sh
sudo systemctl daemon-reload
sudo systemctl enable local_dns.service
sudo systemctl start local_dns.service
```

Retrieve the current DNS. Remember the obtained `****.****.****.****`
```sh
resolvectl dns `awk '$2 == "00000000" {print $1}' /proc/net/route` |awk '{print $NF}'
# ****.****.****.****
```

Modify `/etc/systemd/resolved.conf` to call local_dns.
Set the fallback to the current DNS to ensure functionality even if there are issues with local_dns.

```
[Resolve]
DNS=127.0.0.1:2053
FallbackDNS=****.****.****.****
```

Restart systemd-resolved to use local_dns
```sh
sudo systemctl restart systemd-resolved
```

## Reference

- [How to write a DNS server in Go](https://jameshfisher.com/2017/08/04/golang-dns-server/)

