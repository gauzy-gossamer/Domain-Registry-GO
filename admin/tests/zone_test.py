import asyncio
import pytest

from main import app
from zone.schema import Zone
from zone.zone import create_zone
from fastapi.testclient import TestClient

def test_get_zones(temp_db):
    with TestClient(app) as client:
        response = client.get("/zones")#  json=request_data)
    assert response.status_code == 200
    zones = response.json()

def test_create_zone(temp_db):
    request_data = {"fqdn":"example.com", "ex_period_min":12, "ex_period_max":12}
    with TestClient(app) as client:
        response = client.post("/zones",  json=request_data)
        assert response.status_code == 200

        created_zone = response.json()
        assert created_zone["id"] > 0
        assert created_zone["fqdn"] == request_data['fqdn']
        assert created_zone["ex_period_min"] == request_data['ex_period_min']

        response = client.get("/zones/{}".format(created_zone['id']))
        assert response.status_code == 200

        resp_zone = response.json()
        assert resp_zone == created_zone

def test_get_zone(temp_db):
    with TestClient(app) as client:
        response = client.get("/zones/0")
        assert response.status_code == 404

def test_get_zone_pricelist(temp_db):
    request_data = {"fqdn":"example-t.com", "ex_period_min":12, "ex_period_max":12}
    with TestClient(app) as client:
        response = client.post("/zones",  json=request_data)
        assert response.status_code == 200
        created_zone = response.json()

        request_data = {"zoneid":created_zone['id'], "valid_from":"2010-01-01", "price":10, "operation":"CreateDomain"}

        response = client.post("/zones/{}/pricelist".format(created_zone['id']), json=request_data)
        assert response.status_code == 200

        response = client.get("/zones/{}/pricelist".format(created_zone['id']))
        assert response.status_code == 200

        request_data = {"zoneid":created_zone['id'], "valid_from":"2010-01-01", "price":20, "operation":"CreateDomain"}

        response = client.post("/zones/{}/pricelist".format(created_zone['id']), json=request_data)
        assert response.status_code == 200

        response = client.get("/zones/{}/pricelist".format(created_zone['id']))
        assert response.status_code == 200
        price_list_val = response.json()
        assert price_list_val[0]['price'] == request_data['price']

def test_zone_soa(temp_db):
    request_data = {"fqdn":"example-soa.com", "ex_period_min":12, "ex_period_max":12}
    with TestClient(app) as client:
        response = client.post("/zones",  json=request_data)
        assert response.status_code == 200
        created_zone = response.json()

        request_data = {"ttl":8600, "serial":100, "refresh":1, "update_retr":1, "expiry":1, "minimum":1, "hostmaster":"admin.example-soa.com", "ns_fqdn":"ns.example-soa.com"}

        response = client.post("/zones/{}/soa".format(created_zone['id']), json=request_data)
        assert response.status_code == 200, response

        response = client.get("/zones/{}/soa".format(created_zone['id']))
        assert response.status_code == 200
        zone_soa_val = response.json()
        assert zone_soa_val['ttl'] == request_data['ttl']
        assert zone_soa_val['hostmaster'] == request_data['hostmaster']

def test_zone_ns(temp_db):
    request_data = {"fqdn":"example-ns.com", "ex_period_min":12, "ex_period_max":12}
    with TestClient(app) as client:
        response = client.post("/zones",  json=request_data)
        assert response.status_code == 200
        created_zone = response.json()

        request_data = {"fqdn":"ns.example-soa.com", "addrs":["127.0.0.1"]}

        response = client.post("/zones/{}/ns".format(created_zone['id']), json=request_data)
        assert response.status_code == 200, response

        response = client.get("/zones/{}/ns".format(created_zone['id']))
        assert response.status_code == 200
        zone_ns_val = response.json()
        assert zone_ns_val[0]['fqdn'] == request_data['fqdn']

        response = client.request("delete", "/zones/{}/ns".format(created_zone['id']), json=request_data)
        assert response.status_code == 200, response

        response = client.get("/zones/{}/ns".format(created_zone['id']))
        assert response.status_code == 200
        zone_ns_val = response.json()
        assert len(zone_ns_val) == 0

def test_zone_domain_checks(temp_db):
    request_data = {"fqdn":"example-dcheck.com", "ex_period_min":12, "ex_period_max":12}
    with TestClient(app) as client:
        response = client.post("/zones",  json=request_data)
        assert response.status_code == 200
        created_zone = response.json()

        request_data = {'name':'nonexistant-checker'}
        response = client.post("/domaincheckers/{}".format(created_zone['id']),  json=request_data)
        assert response.status_code != 200, response

        response = client.get("/domaincheckers")
        assert response.status_code == 200, response
        checkers = response.json()
        request_data = {'name':checkers[0]['name']}

        response = client.post("/domaincheckers/{}".format(created_zone['id']),  json=request_data)
        assert response.status_code == 200, response

        response = client.request("delete", "/domaincheckers/{}".format(created_zone['id']), json=request_data)
        assert response.status_code == 200, response
