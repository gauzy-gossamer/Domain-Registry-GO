import os

import pytest

os.environ['TESTING'] = 'True'

from models import database

@pytest.fixture(scope="module")
def temp_db():
    yield database.database 
