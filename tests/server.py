from flask import Flask

app = Flask(__name__)

@app.route('/')
def index():
    return 'Hello World'

@app.route('/vid/<ids>')
def vid(ids):
    return 'video: ' + ids

@app.route('/vid/about')
def vid_about():
    return 'Video about page'

app.run(debug=True, port=7777)