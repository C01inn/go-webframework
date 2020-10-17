from flask import Flask, render_template

app = Flask(__name__, static_folder="tests/static")

@app.route('/')
def index():
    return 'Hello World'

@app.route('/vid/<ids>')
def vid(ids):
    return 'video: ' + ids

@app.route('/vid/<idd>/<name>')
def vid_about(idd, name):
    return 'Video about page \n' + idd + '\n' + name

@app.route('/vid/about/<idd>')
def num22(idd):
    return render_template("index.html")

app.run(debug=True, port=7777)