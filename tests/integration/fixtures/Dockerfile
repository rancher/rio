FROM nginx:alpine
# Pass a build arg
ARG VERSION
RUN /bin/sh -c 'echo hello world: $VERSION' > /usr/share/nginx/html/index.html
