import subprocess
import flask
import os

app = flask.Flask(__name__)

@app.route("/", methods=["GET"])
def challenge():
    arg = flask.request.args.get("top-path", "/")
    command = f"ls -l {arg}"

    print(f"DEBUG: {command=}")
    result = subprocess.run(
        command,
        shell=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        encoding="latin",
    ).stdout

    return f"""
        <html><body>
        Welcome to the dirlister service! Please choose a directory to list the files of:
        <form action="/"><input type=text name=top-path><input type=submit value=Submit></form>
        <hr>
        <b>Output of {command}:</b><br>
        <pre>{result}</pre>
        </body></html>
        """

@app.route("/download", methods=["GET"])
def download():
    filename = flask.request.args.get("file", "")

    if filename != "relatorio-de-teste-de-segurança-WSS.pdf":
        return "No access to download this file", 403

    # Serve the actual PDF file from root directory
    pdf_path = "/relatorio-de-teste-de-segurança-WSS.pdf"

    try:
        return flask.send_file(
            pdf_path,
            as_attachment=True,
            download_name="relatorio-de-teste-de-segurança-WSS.pdf",
            mimetype="application/pdf"
        )
    except FileNotFoundError:
        return "File not found", 404

app.secret_key = os.urandom(8)
app.run(host="0.0.0.0", port=80)

