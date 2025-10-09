FROM python:3.13-slim

WORKDIR /app

COPY server.py .

RUN echo 'WSS{flag88214B93CB66C6AA015B8486E3E65CBA}' > /flag

RUN pip install flask

EXPOSE 8080

CMD ["python", "server.py"]

