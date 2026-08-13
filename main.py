
import requests
import numpy as np
import pandas as pd 
import matplotlib.pyplot as plt
from hmmlearn.hmm import GuassianHMM

API_KEY = 'PKVWY05G2NNY0LQA'

SYMBOL_IBM = 'IBM'
INTERVAL_5MIN = '5min'

# replace the "demo" apikey below with your own key from https://www.alphavantage.co/support/#api-key
url = f'https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol={SYMBOL_IBM}&outputsize=full&apikey={API_KEY}'
r = requests.get(url)
data = r.json()

print(data)