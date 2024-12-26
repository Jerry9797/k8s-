# k8s-webhook

## 创建 CA 证书

### 下载 cfssl 工具

首先，下载 `cfssl` 相关工具到你喜欢的目录：

- `cfssljson_1.6.3_linux_amd64`
- `cfssl-certinfo_1.6.3_linux_amd64`
- `cfssl_1.6.3_linux_amd64`

可以从 [cfssl releases](https://github.com/cloudflare/cfssl/releases) 页面下载。

### 设置工具为可执行文件

下载后，使用以下命令设置为可执行文件：

```
chmod +x *
```
### 移动工具到指定目录
将工具移动到 /usr/cfssl 目录：
```
mv cfssl_1.6.3_linux_amd64 /usr/cfssl
mv cfssl-certinfo_1.6.3_linux_amd64 /usr/cfssl
mv cfssljson_1.6.3_linux_amd64 /usr/cfssl
```
将 /usr/cfssl 目录加入到 PATH 环境变量中，以便可以全局使用：
```
sudo vi /etc/profile
export PATH=$PATH:/usr/cfssl
source /etc/profile
```
### 创建文件和配置
在指定目录（例如 /home/heyilu/hook/certs）创建必要的配置文件。
```
cat > ca-config.json << EOF
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "server": {
        "usages": ["signing"],
        "expiry": "8760h"
      }
    }
  }
}
EOF
```
创建证书请求文件
```
cat > ca-csr.json << EOF
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "zh",
      "L": "bj",
      "O": "bj",
      "OU": "CA"
    }
  ]
}
EOF
```
#### 生成证书
```
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```
- 得到 ca-key.pem  ca.pem
### 服务端证书cs
```
 cat > server-csr.json << EOF
{
  "CN": "admission",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "zh",
      "L": "bj",
      "O": "bj",
      "OU": "bj"
    }
  ]
}
EOF
```
- 签发证书
```
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname=myhook.kube-system.svc \
  -profile=server \
  server-csr.json | cfssljson -bare server
  ```
- 得到 server-key.pem  server.pem

### 取出 yaml/admconfig.yaml 中所需的 caBundle的内容
```
cat ca.pem | base64
```
- 创建secret, 用于[deploy.yaml](yaml/deploy.yaml) 通过volumes挂载
```
  kubectl create secret tls myhook --cert=server.pem --key=server-key.pem  -n kube-system
  ```

## 部署webhook

###  创建符合规则的 pod 时，打上自定义 patch
- 参考 [pods.go](lib/pods.go) 中
  ```
  reviewResponse.Patch = patchImage()
  ```

### 执行 build.sh 得到 myhook可执行文件
```
sh build.sh
```
- 上传 [yaml](yaml) 
- 上传 myhook可执行文件至 /home/heyilu/app
    - 这个目录为[deploy.yaml](yaml/deploy.yaml) 中的挂载目录，目的是将我们的可执行文件myhook挂载进去

- 执行
  ```
  kubectl apply -f deploy.yaml
  kubectl apply -f admconfig.yaml
  kubectl apply -f newpod.yaml
  ```
- 通过调整[newpod.yaml](yaml/newpod.yaml) 中的 metadata.name 来验证是否成功
- 通过给namespace增加label，验证[admconfig.yaml](yaml/admconfig.yaml)中配置的规则是否生效
  ```
  kubectl label namespace default heyilu=enabled
  ```