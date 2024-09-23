# local_dns

開発用に使うローカルDNS。webapiでドメイン名の追加削除を行う  
`/etc/hosts`や`dnsmasq`は`sudo`が必要ですが、このツールは`sudo`なしで使えるのが特徴

## 作成方法
```sh
go build
```

## かんたんな使い方

以下コマンドで起動する

```sh
./local_dns
```
2053ポートでDNSを、2080ポートでwebapiを受け付ける

` --dns-port`, `--http-port` パラメータでポート変更が可能


## ドメイン名の追加例


test.example.comをローカルホストに設定する場合、以下のようにWEBAPIを使う


```sh
curl -POST http://localhost:2080/api --json '{"test.example.com":"127.0.0.1"}'
```

動作確認
```sh
dig @localhost -p 2053 test.example.com
#:
#test.example.com. 60 IN A 127.0.0.1
#:
```

## コマンド

### serviceの起動

`local_dns`を起動する時にパラメータを指定すると、動作を変更できます。

```sh
# dnsを53ポート、 httpを80ポートで開く
./local_dns --dns-port 53 --http-port 80

#同様に環境変数からも設定可能
LOCALDNS_DNS_PORT=53 LOCALDNS_HTTP_PORT=80 ./local_dns

# fallback用のDNSを設定
./local_dns --fallback-ip 8.8.8.8

```


| パラメータ       | 環境変数                | デフォルト | 概要                                        |
|:-----------------|:------------------------|:-----------|:-------------------------------------------|
| `--dns-port`     | `LOCALDNS_DNS_PORT`     | 2053       | DNSのポート                                 |
| `--http-port`    | `LOCALDNS_HTTP_PORT`    | 2080       | HTTPのポート                                |
| `--localhost-only` | `LOCALDNS_LOCALHOST_ONLY` | true   | 127.0.0.1 以外でWebAPIを禁止する            |
| `--fallback-ip`  | `LOCALDNS_FALLBACK_IP`  | n/a        | ドメイン名を持たない場合に代わりに聞くサーバ |
| `--DEBUG`        | `LOCALDNS_DEBUG`        | false      | デバッグ用のログを出す                      |


### webapi

webapiを使ってドメイン名の追加、削除、確認ができます

```sh
# local_dns が持つ domainの一覧を取得
curl http://localohost:2080/api

# 新しいdomainを設定する
curl -POST http://localhost:2080/api -d '{"hoge":192.168.1.2"}'

# domainを削除する。ドメイン名、ip-addressどちらでも良い
curl -DELETE http://localhost:2080/api?domain=hoge
curl -DELETE http://localhost:2080/api?address=192.168.1.2


# すべてのdomainを削除
curl -DELETE http://localhost:2080/api

```


## minikubeでnginxを起動させるときの例

かんたんな使用例としてminikubeで建てたenginxサーバーをブラウザからww.example.comでアクセスさせる例を示す

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

以下のようにマシンの実行ディレクトリにインストールする

```sh
go build
sudo mv local_dns /usr/local/bin/
sudo chown nobody:nogroup /usr/local/bin/local_dns
```

### ubuntu22.04, 24.04 などの設定方法(systemd-resolved)

systemctlに登録、起動するようにする

以下のように `/etc/systemd/system/local_dns.service` を作成する
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

local_dns.serviceを起動する
```sh
sudo systemctl daemon-reload
sudo systemctl enable local_dns.service
sudo systemctl start local_dns.service
```

現状のDNSを取得する。得た `****.****.****.****` を記憶すること

```sh
resolvectl dns `awk '$2 == "00000000" {print $1}' /proc/net/route` |awk '{print $NF}'
# ****.****.****.****
```

`/etc/systemd/resolved.conf` を変更して local_dns を呼ぶようにする
そして、フォールバックを 現状のDNSにすることで local_dns 問題があっても動くようにする
```
[Resolve]
DNS=127.0.0.1:2053
FallbackDNS=****.****.****.****
```

systemd-resolvedを再起動して、 local_dnsを使うようにする
```sh
sudo systemctl restart systemd-resolved
```


## 参考

- [How to write a DNS server in Go](https://jameshfisher.com/2017/08/04/golang-dns-server/)
