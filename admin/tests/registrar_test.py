import pytest
import pydantic

from main import app
from fastapi.testclient import TestClient
from registrar.schema import RegistrarAcl

def test_validation(temp_db):
    examples = (
        ({"registrarid":1, "id":1, "cert":'D1:0C:CC:05:3A:CA:C0:1B:EC:5F:4F:81:0D:C7:99', 'password':'password'}, False),
        ({"registrarid":1, "id":1, "cert":'AA:D1:0C:CC:05:3A:CA:C0:1B:EC:5F:4F:81:0D:C7:99', 'password':'password'}, True),
        ({"id":1, "cert":'AA:D1:0C:CC:05:3A:CA:C0:1B:EC:5F:4F:81:0D:C7:99', 'password':'password'}, False),
    )

    for example, expected in examples:
        try:
            regacl = RegistrarAcl(**example)
            assert expected == True
        except pydantic.error_wrappers.ValidationError:
            assert expected == False

def test_get_registrars(temp_db):
    with TestClient(app) as client:
        response = client.get("/registrars")
    assert response.status_code == 200
    registrars = response.json()

def test_create_registrar(temp_db):
    request_data = {"handle":"TT1-REG"}
    with TestClient(app) as client:
        response = client.post("/registrars",  json=request_data)
        assert response.status_code == 200

        created_reg = response.json()
        assert created_reg["id"] > 0
        assert created_reg["handle"] == request_data['handle']
        assert created_reg["system"] == request_data['system']

        response = client.get("/registrars/{}".format(created_reg['id']))
        assert response.status_code == 200

        resp_reg = response.json()
        assert created_reg["id"] == resp_reg['id']

def test_create_registrar_acl(temp_db):
    request_data = {"handle":"TEST-REG","system":False}
    with TestClient(app) as client:
        response = client.post("/registrars",  json=request_data)
        assert response.status_code == 200

        created_reg = response.json()
        assert created_reg["id"] > 0
        assert created_reg["handle"] == request_data['handle']
        assert created_reg["system"] == request_data['system']

        request_data = {"cert":'49:D1:0C:CC:05:3A:CA:C0:1B:EC:5F:4F:81:0D:C7:99', 'password':'password'}

        response = client.post("/registrars/{}/acl".format(created_reg['id']), json=request_data)
        assert response.status_code == 200
        resp_acl = response.json()
        assert resp_acl['registrarid'] == created_reg['id']

        request_data = {"cert":'59:D1:0C:CC:05:3A:CA:C0:1B:EC:5F:4F:81:0D:C7:99', 'password':'password'}

        response = client.put("/registrars/{}/acl".format(created_reg['id']), json=request_data)
        assert response.status_code == 200
        resp_acl = response.json()
        assert resp_acl['cert'] == request_data['cert']

        request_data = [
            {'ipaddr':'127.0.0.1'},
            {'ipaddr':'127.0.0.2'},
        ]

        response = client.put("/registrars/{}/ips".format(created_reg['id']), json=request_data)
        assert response.status_code == 200

        response = client.get("/registrars/{}/ips".format(created_reg['id']))
        assert response.status_code == 200
        resp_ips = response.json()
        assert len(resp_ips) == 2

        request_data = [
            {'ipaddr':'127.0.0.1'},
        ]

        response = client.put("/registrars/{}/ips".format(created_reg['id']), json=request_data)
        assert response.status_code == 200

        response = client.get("/registrars/{}/ips".format(created_reg['id']))
        assert response.status_code == 200
        resp_ips = response.json()
        assert len(resp_ips) == 1

def test_get_registrar(temp_db):
    with TestClient(app) as client:
        response = client.get("/registrars/0")
        assert response.status_code == 404

def test_get_registrar_acl(temp_db):
    with TestClient(app) as client:
        response = client.get("/registrars/0/acl")
        assert response.status_code == 404
