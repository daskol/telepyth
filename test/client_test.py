#   encoding: utf8
#   client_test.py
#   TODO: it could be broken

import io
import telepyth.client as client
import matplotlib.pyplot as plt


fig = plt.figure()
ax = fig.add_subplot(1, 1, 1)
ax.plot([0, 1], [0, 1], '.-k')

s = io.StringIO()
s.write('test')
s.seek(0)

cli = client.TelePythClient(token='6768929855443357532', debug=True)

print(cli(s))
print(cli.send_figure(fig, 'Title'))
