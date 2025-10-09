#!/bin/sh

# Adiciona o host MinIO
/usr/bin/mc alias set myminio $MINIO_SERVER_HOST $MINIO_ACCESS_KEY $MINIO_SECRET_KEY

# Tenta criar o bucket (o -p evita erro se já existir)
echo "Criando o bucket '$BUCKET_NAME'..."
/usr/bin/mc mb myminio/$BUCKET_NAME -p

# Copia as imagens da pasta local mapeada para o bucket
echo "Copiando imagens de '$LOCAL_IMAGE_FOLDER' para o bucket '$BUCKET_NAME'..."
# O --recursive garante que todos os arquivos e subpastas sejam copiados
/usr/bin/mc cp --recursive $LOCAL_IMAGE_FOLDER/ myminio/$BUCKET_NAME/

echo "Setup do MinIO concluído!"

# O contêiner mc-setup pode sair após a conclusão
exit 0