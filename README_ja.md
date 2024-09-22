
# local_dns

開発用に使うローカルDNS。webapiでドメイン名の追加削除を行う  
/etc/hostsや dnsmasq と同等だがwebapiで操作できることを特徴とする

## 作成方法
```sh
go build
```

## かんたんな使い方
```sh
./local_dns
```
2053ポートでDNSを、2080ポートでwebapiを受け付ける

## ドメイン名の追加例
```sh
curl -POST http://localhost:2080/api --json '{"test.example.com.":"127.0.0.1"}'
```

## 動作確認
```sh
dig @localhost:2053 test.example.com
```
```
test.example.com. 60 IN A 127.0.0.1
```

## minikubeでnginxを起動させるときの例

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

# regist minikube ip as www.example.com 
curl -POST http://localhost:2080/api --json "{\"www.example.com.\":\"$(minikube ip)\"}"
```


## 開発マシンへの統合

### ubuntu22.04, 24.04 などの設定方法(systemd-resolved)

```sh
go build
sudo mv local_dns /usr/local/bin/
sudo chown nobody:nogroup /usr/local/bin/local_dns
```

#### systemctlに登録、起動するようにする
`/etc/systemd/system/local_dns.service`
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

#### 現状のDNSを取得
```sh
resolvectl dns `awk '$2 == "00000000" {print $1}' /proc/net/route` |awk '{print $NF}'
```
```
****.****.****.****
```

#### `/etc/systemd/resolved.conf` を変更
```
DNS=127.0.0.1:2053
FallbackDNS=****.****.****.****
```

#### systemd-resolvedを再起動
```sh
sudo systemctl restart systemd-resolved
```

## コマンド例
```sh
curl http://172.18.0.2/api
curl -POST http://172.18.0.2/api -d '{"hoge.":"127.0.0.1"}'
dig @172.18.0.2 hoge
curl -DELETE http://172.18.0.2/api -d '["hoge"]'
```

## 参考
[https://jameshfisher.com/2017/08/04/golang-dns-server/](https://jameshfisher.com/2017/08/04/golang-dns-server/)
