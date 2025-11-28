FROM node:24 as build
WORKDIR /frontend
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build


FROM caddy:2
WORKDIR /frontend
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /frontend/dist ./

EXPOSE 80
