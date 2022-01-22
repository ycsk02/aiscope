FROM ubuntu:20.04
LABEL maintainer="SuKai"
ARG MLFLOW_VERSION=1.21.0

WORKDIR /mlflow/
RUN apt update -y && apt install -y python3-pip
RUN pip install mlflow==$MLFLOW_VERSION boto3==1.19.12 pymysql -i https://pypi.tuna.tsinghua.edu.cn/simple
EXPOSE 5000

ENV BACKEND_URI sqlite:////mlflow/mlflow.db

CMD mlflow server --backend-store-uri ${BACKEND_URI} --default-artifact-root ${ARTIFACT_ROOT} --host 0.0.0.0 --port 5000