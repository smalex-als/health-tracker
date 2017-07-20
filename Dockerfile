FROM golang:1.7

RUN wget https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-163.0.0-linux-x86_64.tar.gz
RUN tar xfz google-cloud-sdk-163.0.0-linux-x86_64.tar.gz
RUN ./google-cloud-sdk/install.sh
RUN ./google-cloud-sdk/bin/gcloud components install app-engine-go

ENV PATH /go/google-cloud-sdk/bin:$PATH
