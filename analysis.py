import json
import datetime
import dateutil.parser
import numpy as np
import matplotlib.pyplot as plt

class Pause:
	def __init__(self, start, end):
		self.start = dateutil.parser.parse(start) if start != None else None
		self.end = dateutil.parser.parse(end) if end != None else None

def parse_pauses(data):
	ret = []
	for d in data: ret.append(Pause(d.get('start'), d.get('end')))
	return ret

class WorkUnit:
	def __init__(self, start, end, target_length, activity, activity_category, complete, pause_data):
		self.start = dateutil.parser.parse(start) if start != None else None
		self.end = dateutil.parser.parse(end) if end != None else None
		self.target_length = target_length
		self.activity = activity
		self.activity_category = activity_category
		self.complete = complete
		self.pauses = parse_pauses(pause_data) if pause_data != None else []

def get_works(filename):
	with open(filename, "r") as f:
		data = json.load(f)
		return data

def parse_data():
	data = get_works('log.json')

	ret = {}
	for d in data:
		w = WorkUnit(d['start'], d['end'], d['target_length'], d['activity'], d['activity_category'], d['complete'], d['pauses'])

		if w.start.date() in ret:
			ret[w.start.date()].append(w)
		else:
			ret[w.start.date()] = [w]

	return ret

def plot():
	data = parse_data()
	ind = np.arange(len(data))
	width = 0.35
	plot_data = [len(y) for x, y in data.items()]
	# print(matplotlib.rcParams['backend'])
	p1 = plt.bar(ind, plot_data, width)
	plt.xticks(ind, [x for x, _ in data.items()])

	plt.plot(range(10))
	plt.show()

plot()