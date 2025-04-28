FROM ezhttp/ezhttp:0.0.3

USER root
RUN rm -r public && rm config.json
COPY ./config.json .
COPY ./public ./public
USER appuser:appgroup
