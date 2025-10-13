import subprocess
import flask
import os
from datetime import datetime

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
        <!DOCTYPE html>
        <html lang="pt-BR">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>WSS - Sistema de Gerenciamento de Arquivos</title>
            <style>
                * {{
                    margin: 0;
                    padding: 0;
                    box-sizing: border-box;
                }}
                body {{
                    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    min-height: 100vh;
                    padding: 20px;
                }}
                .container {{
                    max-width: 1200px;
                    margin: 0 auto;
                }}
                .header {{
                    background: white;
                    padding: 20px 30px;
                    border-radius: 10px 10px 0 0;
                    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                }}
                .header h1 {{
                    color: #333;
                    font-size: 24px;
                    display: flex;
                    align-items: center;
                    gap: 10px;
                }}
                .logo {{
                    width: 40px;
                    height: 40px;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    border-radius: 8px;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    color: white;
                    font-weight: bold;
                    font-size: 20px;
                }}
                .subtitle {{
                    color: #666;
                    font-size: 14px;
                    margin-top: 5px;
                }}
                .main-content {{
                    background: white;
                    padding: 30px;
                    border-radius: 0 0 10px 10px;
                    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                }}
                .search-section {{
                    margin-bottom: 30px;
                }}
                .search-section h2 {{
                    color: #333;
                    font-size: 18px;
                    margin-bottom: 15px;
                }}
                .search-form {{
                    display: flex;
                    gap: 10px;
                }}
                .search-input {{
                    flex: 1;
                    padding: 12px 15px;
                    border: 2px solid #e0e0e0;
                    border-radius: 8px;
                    font-size: 14px;
                    transition: border-color 0.3s;
                }}
                .search-input:focus {{
                    outline: none;
                    border-color: #667eea;
                }}
                .btn {{
                    padding: 12px 30px;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    color: white;
                    border: none;
                    border-radius: 8px;
                    font-size: 14px;
                    font-weight: 600;
                    cursor: pointer;
                    transition: transform 0.2s, box-shadow 0.2s;
                }}
                .btn:hover {{
                    transform: translateY(-2px);
                    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
                }}
                .results-section {{
                    margin-top: 20px;
                }}
                .results-header {{
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 15px;
                    padding-bottom: 10px;
                    border-bottom: 2px solid #e0e0e0;
                }}
                .results-header h3 {{
                    color: #333;
                    font-size: 16px;
                }}
                .results-info {{
                    color: #666;
                    font-size: 13px;
                }}
                .output-box {{
                    background: #f8f9fa;
                    border: 1px solid #e0e0e0;
                    border-radius: 8px;
                    padding: 20px;
                    font-family: 'Courier New', monospace;
                    font-size: 13px;
                    line-height: 1.6;
                    overflow-x: auto;
                    color: #333;
                    white-space: pre-wrap;
                }}
                .footer {{
                    text-align: center;
                    margin-top: 20px;
                    color: white;
                    font-size: 13px;
                }}
                .info-badge {{
                    display: inline-block;
                    background: #e3f2fd;
                    color: #1976d2;
                    padding: 8px 15px;
                    border-radius: 20px;
                    font-size: 12px;
                    margin-top: 15px;
                }}
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>
                        <div class="logo">TC</div>
                        Sistema de Gerenciamento de Arquivos
                    </h1>
                    <div class="subtitle">Versão 2.4.1 | Ambiente de Produção</div>
                </div>
                <div class="main-content">
                    <div class="search-section">
                        <h2>📁 Explorador de Diretórios</h2>
                        <form action="/" method="GET" class="search-form">
                            <input type="text" name="top-path" value="{arg}"
                                   placeholder="Digite o caminho do diretório (ex: /home, /var/log)"
                                   class="search-input">
                            <button type="submit" class="btn">Listar Arquivos</button>
                        </form>
                        <div class="info-badge">
                            💡 Dica: Use caminhos absolutos para melhores resultados
                        </div>
                    </div>
                    <div class="results-section">
                        <div class="results-header">
                            <h3>Resultados da Busca</h3>
                            <div class="results-info">Diretório: {arg}</div>
                        </div>
                        <div class="output-box">{result if result else 'Nenhum resultado encontrado.'}</div>
                    </div>
                </div>
                <div class="footer">
                    © 2025 TechCorp | Todos os direitos reservados | Sistema Interno
                </div>
            </div>
        </body>
        </html>
        """

@app.route("/about", methods=["GET"])
def about():
    return """
        <!DOCTYPE html>
        <html lang="pt-BR">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Sobre - WSS Sistema</title>
            <style>
                * {{
                    margin: 0;
                    padding: 0;
                    box-sizing: border-box;
                }}
                body {{
                    font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    min-height: 100vh;
                    padding: 20px;
                }}
                .container {{
                    max-width: 800px;
                    margin: 0 auto;
                    background: white;
                    padding: 40px;
                    border-radius: 10px;
                    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
                }}
                h1 {{
                    color: #333;
                    margin-bottom: 20px;
                }}
                p {{
                    color: #666;
                    line-height: 1.8;
                    margin-bottom: 15px;
                }}
                .back-link {{
                    display: inline-block;
                    margin-top: 20px;
                    color: #667eea;
                    text-decoration: none;
                    font-weight: 600;
                }}
                .back-link:hover {{
                    text-decoration: underline;
                }}
            </style>
        </head>
        <body>
            <div class="container">
                <h1>Sobre o Sistema</h1>
                <p>O Sistema de Gerenciamento de Arquivos WSS é uma ferramenta interna desenvolvida para auxiliar administradores de sistema na navegação e listagem de diretórios.</p>
                <p><strong>Versão:</strong> 2.4.1</p>
                <p><strong>Última Atualização:</strong> Janeiro 2025</p>
                <p><strong>Desenvolvido por:</strong> Departamento de TI - WSS Corporation</p>
                <a href="/" class="back-link">← Voltar ao Sistema</a>
            </div>
        </body>
        </html>
    """

@app.route("/download", methods=["GET"])
def download():
    filename = flask.request.args.get("file", "")

    if filename != "relatorio-de-teste-WSS.pdf":
        return f"Acesso negado ao arquivo '{filename}'. Tente com um parâmetro 'file' diferente.", 403

    # Serve the actual PDF file from root directory
    pdf_path = "/relatorio-de-teste-WSS.pdf"

    try:
        return flask.send_file(
            pdf_path,
            as_attachment=True,
            download_name="relatorio-de-teste-WSS.pdf",
            mimetype="application/pdf"
        )
    except FileNotFoundError:
        return "File not found", 404

app.secret_key = os.urandom(8)
app.run(host="0.0.0.0", port=80)

