FROM caddy:2

COPY index.html /usr/share/caddy/

EXPOSE 80

CMD ["caddy", "file-server", "-r", "/usr/share/caddy/"]
