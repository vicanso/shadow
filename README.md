# shadow

用于生成图片的模糊阴影

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