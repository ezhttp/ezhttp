FROM alpine:3.21.2

RUN apk add --no-cache curl && \
	addgroup -S appgroup && \
	adduser -S appuser -G appgroup --home "/usr/src/app" --no-create-home

WORKDIR /usr/src/app

COPY ./bin/ezhttp.alpine.bin ./ezhttp
RUN chmod +x ./ezhttp
COPY ./config.json .
COPY ./public ./public

USER appuser:appgroup

# Docker only
#HEALTHCHECK --interval=15s --retries=2 --start-period=5s --timeout=5s CMD curl --fail http://127.0.0.1:8080/health || exit 1

EXPOSE 8080
CMD ["/usr/src/app/ezhttp"]
