import struct
import base64
import hashlib

# use dnssec-keygen to generate DNSKEYs, e.g.:
# dnssec-keygen -a RSASHA256 -b 1024 domain.ex.com
class DSRecord():
    def __calc_keytag(self, flags, protocol, algorithm, dnskey):
        st = struct.pack('!HBB', int(flags), int(protocol), int(algorithm))
        st += base64.b64decode(dnskey)

        cnt = 0
        for idx in range(len(st)):
            s = struct.unpack('B', st[idx:idx+1])[0]
            if (idx % 2) == 0:
                cnt += s << 8
            else:
                cnt += s

        return ((cnt & 0xFFFF) + (cnt >> 16)) & 0xFFFF

    def __calc_ds(self, domain: str, flags: int, protocol: int, algorithm: int, dnskey: str, digest_algo: int):
        if domain.endswith('.') is False:
            domain += '.'

        signature = bytes()
        for i in domain.split('.'):
            signature += struct.pack('B', len(i)) + i.encode()

        signature += struct.pack('!HBB', int(flags), int(protocol), int(algorithm))
        signature += base64.b64decode(dnskey)

        if digest_algo == 1:
            # sha1 
            return hashlib.sha1(signature).hexdigest().upper()
        else:
            # use sha256
            return hashlib.sha256(signature).hexdigest().upper()

    def generate_dsrecord(self, domain: str, key:str, flags: int = 257, protocol: int = 3, algorithm: int = 12, digest_algo: int = 3):
        keyid = self.__calc_keytag(flags, protocol, algorithm, key)
        digest = self.__calc_ds(domain, flags, protocol, algorithm, key, digest_algo)

        return {
            'type':'ds', 
            'data':{
                    'keyTag':keyid, 'alg':algorithm, 'digestType':digest_algo, 'digest':digest.upper(),
                    'keyData':{'type':'key','data':{
                        'flags':flags, 'protocol':protocol, 'alg':algorithm, 'pubKey':key
                    }
                }
            }
        }

