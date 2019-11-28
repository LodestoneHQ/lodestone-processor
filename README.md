# lodestone-processor

Tasks

- Needs to process entries from Redis
    - Processing via TIKA integration
        - output stored in Elasticsearch
    - Thumbnail generation
        - output stored in minio storage

# Local Development

```
docker build --tag=lodestone-processor .
docker run -v `pwd`:/go/src/github.com/analogj/lodestone-processor/ lodestone-processor

```
