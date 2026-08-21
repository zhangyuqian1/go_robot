# 阶段 1：构建 Go 应用程序
FROM harbor.qa.dobest.com/common/ubuntu:24.04_fix


LABEL authors="houyuqi@dobest.com"

# 设置容器内的当前工作目录
WORKDIR /home/yoka/

# 复制 Go 模块文件和源代码
ENV GOPROXY https://goproxy.cn

ADD . .

COPY ./entrypoint.sh /usr/bin/entrypoint.sh

RUN chmod +x ./stressTest && \
    chmod +x /usr/bin/entrypoint.sh

ENTRYPOINT ["/usr/bin/entrypoint.sh"]