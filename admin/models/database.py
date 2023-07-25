import databases

from environs import Env

env = Env()
env.read_env()

DB_USER = env("DBUSER")
DB_PASSWORD = env("DBPASSWORD")
DB_HOST = env("DBHOST")
DB_NAME = env("DBNAME")
DB_PORT = env("DBPORT")

TEST_SQLALCHEMY_DATABASE_URL = (
    f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
)
database = databases.Database(TEST_SQLALCHEMY_DATABASE_URL)

TORTOISE_DATABASE_URL = (
    f"postgres://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
)