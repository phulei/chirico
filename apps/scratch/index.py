import os
from google.appengine.ext.webapp import template
from google.appengine.ext import db
from google.appengine.api import mail
import cgi
import wsgiref.handlers

from google.appengine.api import users
from google.appengine.ext import webapp
# This is URL fetch
from google.appengine.api import urlfetch


class DB(db.Model):
  author = db.UserProperty()
  content = db.StringProperty(multiline=True)
  greetings = db.StringProperty(multiline=True)
  url_content = db.StringProperty(multiline=True)
  date = db.DateTimeProperty(auto_now_add=True)

class MainPage(webapp.RequestHandler):
  def get(self):
    conn = DB.all().order('-date')
    results = conn.fetch(10)


    if users.get_current_user():
      url = users.create_logout_url(self.request.uri)
      url_linktext = 'Logout'
    else:
      url = users.create_login_url(self.request.uri)
      url_linktext = 'Login'

      greetings='Sample Stuff'  

    template_values = {
      'results': results,
      'url': url,
      'url_linktext': url_linktext,
      }

    path = os.path.join(os.path.dirname(__file__), 'index.html')
    self.response.out.write(template.render(path, template_values))




class Guestbook(webapp.RequestHandler):
  def post(self):
    db = DB()


    url = "https://mchirico.org/"
    url_result = urlfetch.fetch(url)
    if url_result.status_code == 200:
      m=url_result.content


    if users.get_current_user():
      db.author = users.get_current_user()

    db.content = self.request.get('content')
    db.url_content = m
    db.put()
    self.redirect('/')


class Action0(webapp.RequestHandler):
  def post(self):
    greeting = Greeting()

    if users.get_current_user():
      greeting.author = users.get_current_user()

    greeting.content = "Action 0"
    greeting.put()

    message = mail.EmailMessage(sender="mchirico@gmail.com",
              subject="Your account has been approved")
    message.to = "Mike Chirico <mchirico@cwxstat.org>"
    message.body = """
    This is a sample body
"""
    message.send()

    self.redirect('http://www.chirico.org')



class Test(webapp.RequestHandler):
   def post(self):
     mydb = DB()
     mydb.content="stuff"
     mydb.put()
     self.redirect('/')







def main():
  application = webapp.WSGIApplication(
                                       [('/', MainPage),
                                        ('/sign', Guestbook),
                                        ('/Action0', Action0),
                                        ('/test', Test)

                                            ],
                                       debug=True)
  wsgiref.handlers.CGIHandler().run(application)

if __name__ == "__main__":
  main()
