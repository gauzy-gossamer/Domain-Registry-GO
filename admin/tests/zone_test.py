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
        


'''
@pytest.mark.freeze_time("2015-10-21")
def test_user_detail_forbidden_with_expired_token(temp_db, freezer):
    user = UserCreate(
        email="sidious@deathstar.com",
        name="Palpatine",
        password="unicorn"
    )
    with TestClient(app) as client:
        # Create user and use expired token
        loop = asyncio.get_event_loop()
        user_db = loop.run_until_complete(create_user(user))
        freezer.move_to("'2015-11-10'")
        response = client.get(
            "/users/me",
            headers={"Authorization": f"Bearer {user_db['token']['token']}"}
        )
    assert response.status_code == 401
'''
