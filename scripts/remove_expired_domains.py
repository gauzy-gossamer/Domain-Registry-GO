import logging

import grpc
import regrpc.registry_pb2 as regpb2
import regrpc.registry_pb2_grpc as regpb2_grpc

from config import read_config

def run():
    config = read_config()

    # domain removal is processed through grpc, so that all relevant checks are performed,
    # and potentially unlinked objects are removed as well
    with grpc.insecure_channel(config['grpc_host']) as channel:
        stub = regpb2_grpc.RegistryStub(channel)

        response = stub.LoginSystem(regpb2.Empty())
        sessionid = response.sessionid

        for domain in stub.GetExpiredDomains(regpb2.Session(sessionid=sessionid)):
            response = stub.DeleteDomain(regpb2.Domain(sessionid=sessionid, name=domain.name))
            logging.info("Delete return code: {}".format(response.return_code))

        response = stub.LogoutSystem(regpb2.Session(sessionid=sessionid))
        logging.info("Logout return code: {}".format(response.return_code))

if __name__ == '__main__':
    logging.basicConfig()
    run()
