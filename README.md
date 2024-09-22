
# local_dns

Local DNS for development use. Add and delete domain names via web API.  
Similar to /etc/hosts or dnsmasq but operable via web API.

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
dig @localhost:2053 test.example.com
#:
#test.example.com. 60 IN A 127.0.0.1
#:
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
DNS=127.0.0.1:2053
FallbackDNS=****.****.****.****
```

Restart systemd-resolved to use local_dns
```sh
sudo systemctl restart systemd-resolved
```

## Command

### webapi
```sh
# Get a list of domains held by local_dns
curl http://localhost/api

# Set a new domain
curl -POST http://localhost/api -d '{"foo":"127.0.0.1"}'

# Delete a domain. Either the domain name or the IP address can be used
curl -DELETE http://localhost/api -d '["foo"]'



## Reference
[https://jameshfisher.com/2017/08/04/golang-dns-server/](https://jameshfisher.com/2017/08/04/golang-dns-server/)

