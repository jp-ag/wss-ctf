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

    if filename != "report.pdf":
        return "No access to download this file", 403

    # Create the report.pdf content
    content = "hello world"

    response = flask.Response(content)
    response.headers['Content-Disposition'] = 'attachment; filename=report.pdf'
    response.headers['Content-Type'] = 'text/plain'

    return response

app.secret_key = os.urandom(8)
app.run(host="0.0.0.0", port=80)

