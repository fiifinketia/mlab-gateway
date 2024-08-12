FROM devopsfaith/krakend as builder
ARG ENV=dev

COPY krakend.tmpl .
COPY config .

## Save temporary file to /tmp to avoid permission errors
RUN FC_ENABLE=1 \
    FC_OUT=/tmp/krakend.json \
    FC_PARTIALS="/etc/krakend/partials" \
    FC_SETTINGS="/etc/krakend/settings/$ENV" \
    FC_TEMPLATES="/etc/krakend/templates" \
    krakend check -d -t -c krakend.tmpl

# The linting needs the final krakend.json file
RUN krakend check -c /tmp/krakend.json --lint

FROM devopsfaith/krakend
COPY --from=builder --chown=krakend /tmp/krakend.json .
# WORKDIR /etc/krakend/plugins
# RUN make amd64
EXPOSE 8080

# Uncomment with Enterprise image:
# COPY LICENSE /etc/krakend/LICENSE
