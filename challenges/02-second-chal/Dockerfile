FROM python:3.13-slim

WORKDIR /app

COPY server.py .

RUN echo '1234' > /flag

RUN pip install flask

EXPOSE 80

CMD ["python", "server.py"]

