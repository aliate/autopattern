#!/usr/bin/env python

import urllib2
import json

es_url = 'http://172.24.5.91:9200';

def http_get(path):
    return urllib2.urlopen('{}/{}'.format(es_url, path)).read()

def http_post(path, data):
    return urllib2.urlopen(urllib2.Request('{}/{}'.format(es_url, path), data)).read()

def get_project_mapping():
    return [index for index in json.loads(http_get('_mapping')) if not index.startswith('.kibana')]

def get_kibana_index_patterns():
    p = json.loads(http_get('.kibana/index-pattern/_search'))
    return [v["_source"]['title'] for v in p['hits']['hits']]

def create_index_pattern(pattern):
    return json.loads(http_post('.kibana/index-pattern/{}'.format(pattern), json.dumps({'title':pattern})))

if __name__ == "__main__":
    exists = get_kibana_index_patterns()
    mappings = get_project_mapping()

    print("exists: ", exists)
    print("mappings: ", mappings)

    print("\nstart process:\n")
    processed = []
    for m in mappings:
        p = '.'.join(m.split('.')[:2]) + '.*'
        if p in processed:
            continue
        print("{} -> {}".format(m, p))
        if p not in exists:
            print(create_index_pattern(p))
        else:
            print('pattern: {} already exists!'.format(p))
        print('')
        processed.append(p)
