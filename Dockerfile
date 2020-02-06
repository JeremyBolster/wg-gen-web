FROM golang:alpine AS build-back
WORKDIR /app
ADD . .
RUN go build -o wg-gen-web-linux

FROM node:10-alpine AS build-front
WORKDIR /app
ADD ui .
RUN npm install
RUN npm run build

FROM alpine
WORKDIR /app
COPY --from=build-back /app/wg-gen-web-linux .
COPY --from=build-front /app/dist ./ui/dist
ADD .env .
RUN chmod +x ./wg-gen-web-linux
RUN apk add --no-cache ca-certificates
EXPOSE 8080

CMD ["/app/wg-gen-web-linux"]