FROM golang:1.17

RUN apt-get update \
    && apt-get install -y zsh \
    && apt-get -y purge

ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd -s /bin/zsh --uid ${USER_UID} --gid ${USERNAME} -m ${USERNAME}
