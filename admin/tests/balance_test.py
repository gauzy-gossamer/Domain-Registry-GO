import pytest
import pydantic

from main import app
from fastapi.testclient import TestClient

def test_list_balance(temp_db):
    with TestClient(app) as client:
        response = client.get("/balance")
        assert response.status_code == 200
        balance = response.json()

def test_get_balance(temp_db):
    with TestClient(app) as client:
        response = client.get("/balance/0")
        assert response.status_code == 200
        balance = response.json()
        assert len(balance) == 0

        response = client.get("/registrars")
        assert response.status_code == 200
        registrars = response.json()

        if len(registrars) > 0:
            response = client.get("/balance/{}".format(registrars[0]['id']))
            assert response.status_code == 200
            balance = response.json()

def test_add_balance(temp_db):
    with TestClient(app) as client:
        response = client.get("/registrars")
        assert response.status_code == 200
        registrars = response.json()

        assert len(registrars) > 0

        response = client.get("/registrars/{}/zones".format(registrars[0]['id']))
        assert response.status_code == 200
        zones = response.json()
        request_data = {'zone':zones[0]['zone'], 'balance_change':10}
        response = client.post("/balance/{}".format(registrars[0]['id']), json=request_data)
        assert response.status_code == 200, response.json()

def test_list_invoice(temp_db):
    with TestClient(app) as client:
        response = client.get("/invoice")
        assert response.status_code == 200
        balance = response.json()
