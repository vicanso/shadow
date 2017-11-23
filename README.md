# shadow

用于图片的缩放、压缩等处理


curl 'http://127.0.0.1:3015/@images/optim?url=xxx&type=jpeg'

curl 'http://127.0.0.1:3015/@images/resize?url=xxx&type=jpeg&width=300&height=100'


## docker

### docker build

```bash
docker build -t vicanso/shadow .
```

### docker run

```bash
docker run -d --restart=always \
  -p 3015:3015 \
  -v /data/xxx:/covers \
  vicanso/shadow
```