# ---------- مرحلة البناء ----------
FROM golang:1.25-alpine AS build

WORKDIR /src

# التبعيات أولًا حتى تُعاد الاستفادة من طبقة الكاش عند تغيّر الشيفرة فقط.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO معطّل لأن محرّك SQLite المستخدم مكتوب بلغة Go بالكامل،
# فينتج بايناري ساكن يعمل على صورة scratch بلا أي مكتبات نظام.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/surveyapp .

# ---------- صورة التشغيل ----------
FROM scratch

COPY --from=build /out/surveyapp /surveyapp

ENV ADDR=":8080" \
    DB_PATH="/data/survey.db"

VOLUME ["/data"]
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/surveyapp", "-healthcheck"]

ENTRYPOINT ["/surveyapp"]
