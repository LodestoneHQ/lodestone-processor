# lodestone-processor

Tasks

- Needs to process entries from Redis
    - Processing via TIKA integration
        - output stored in Elasticsearch
    - Thumbnail generation
        - output stored in minio storage

# How to build

```bash
docker build -f Dockerfile.document --tag lodestone-document-processor .
docker run lodestone-document-processor

docker build -f Dockerfile.thumbnail --tag lodestone-thumbnail-processor .
docker run lodestone-thumbnail-processor
```
