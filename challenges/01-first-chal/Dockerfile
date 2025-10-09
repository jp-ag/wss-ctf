FROM python:3.13-slim

WORKDIR /app

COPY server.py .
COPY relatorio-de-teste-de-seguranca-WSS.pdf /relatorio-de-teste-de-seguranca-WSS.pdf

RUN echo 'WSS{flag_voce_esta_no_caminho_certo!}' > /flag

RUN pip install flask

EXPOSE 8080

CMD ["python", "server.py"]

