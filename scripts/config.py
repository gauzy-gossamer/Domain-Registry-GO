import sys
import json
import psycopg2

def get_pg_conn(config):
    try:
        db = psycopg2.connect(f"user='{config['db']['user']}' host='{config['db']['host']}' password='{config['db']['password']}' port='{config['db']['port']}'")
    except:
        raise Exception("failed to connect to database")

    return db

def read_config(config_file='config.json'):
    with open(config_file, 'r') as f:
        content = f.read()
        return json.loads(content)
