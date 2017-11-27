# shadow

用于图片的缩放、压缩等处理

公共参数列表：

- `query.url` 源图片地址，在请求的querstring中指定源图片的地址
- `body.base64` 源图片数据，在POST中指定源数据的base64，优先以`query.url`的形式获取数据 
- `query.type` 转换后的图像格式，如果不指定，则与源图片一致，支持`png` `jpeg` `webp`
- `query.output` 转换后返回的数据格式，如果不指定，则以字节形式返回，可指定为`base64`
- `query.quality` 压缩图片选择的质量，`jpeg`与`webp`支持，如果设置为0，`webp`表示使用Lossless:true（对于有透明处理的PNG，使用此方式）

## optim

图片压缩，对指定的图片压缩处理

```bash
curl 'http://127.0.0.1:3015/@images/optim?url=xxx&type=jpeg'
```

## resize

图片尺寸转换，对指定的图片做缩放处理

参数列表：

- `query.width` 转换后的图片宽度，如果不指定，则等比例缩放（高度和宽度必须最少指定其中一个）
- `query.height` 转换后的图片高度，如果不指定，则等比较缩放（高度和宽度必须最少指定其中一个）

```bash
curl 'http://127.0.0.1:3015/@images/resize?url=xxx&type=jpeg&width=300&height=100'
```

## docker

由于`webp`模块需要用到gcc，因此编译执行文件没办法使用指定`GOOS`的方式，因此使用`docker pull golang`的镜像来做构建



### docker build

```bash
docker run -it --rm -v ~/go:/go -v ~/github:/github golang bash
```

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