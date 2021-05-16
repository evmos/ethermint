# This image is not yet distroless because distroless images expect thatthe entrypoint will be a .js file and ours is vue-cli-service
# We use lopsided/archlinux for comptability

FROM lopsided/archlinux
COPY . .

RUN pacman -Syyu --noconfirm npm
RUN npm install
RUN npm run build

EXPOSE 8080
CMD ["/usr/bin/npm","run","serve"]