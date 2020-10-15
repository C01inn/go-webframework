from flask import Flask

app = Flask(__name__)

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
    return idd 

app.run(debug=True, port=7777)