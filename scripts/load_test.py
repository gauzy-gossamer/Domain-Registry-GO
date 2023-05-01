import sys
import random
import time
import string
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

from multiprocessing import Pool

from eppy_ripn import create_client

settings = {
    'registrar': "TEST-REG",
    'passwd':"passwd",
    'test_contact':"TEST-CONTACT6",
    'test_host':'ns1.syst.test.com',
    'zone':'.test.com',
}

def assert_retcode(result, code):
    assert result['epp']['response']['result']['@code'] == code

def id_generator(size=10, chars=string.ascii_lowercase + string.digits):
    return ''.join(random.choice(chars) for _ in range(size))

def test_info(client, generated_domain):
    assert_retcode(client.info_domain(generated_domain), "1000")
    assert_retcode(client.info_host(settings['test_host']), "1000")
    assert_retcode(client.info_domain(generated_domain), "1000")
    return 3

def test_infoupdate(client, generated_domain):
    assert_retcode(client.info_domain(generated_domain), "1000")
    assert_retcode(client.info_host(settings['test_host']), "1000")
    assert_retcode(client.update_domain(generated_domain, description=["hello"]), "1000")
    return 3

def test_createdelete(client, generated_domain):
    generated_domain = id_generator() + settings['zone']
    assert_retcode(client.create_domain(generated_domain, registrant=settings['test_contact']), "1000")
    assert_retcode(client.info_domain(generated_domain), "1000")
    assert_retcode(client.delete_domain(generated_domain), "1000")
    return 3

def run_test(n):
    cert = "client.crt"
    key = "client.key"
    client = create_client(settings['registrar'], settings['passwd'], lang='en',
                    cert=cert, key=key, server='localhost', port=8081, print_commands=False)

    completed = 0
    generated_domain = id_generator() + settings['zone']
    client.create_domain(generated_domain, registrant=settings['test_contact'])

    for i in range(50):
        try:
            completed += test_infoupdate(client, generated_domain)
        except Exception as exc:
            sys.stderr.write('{}\n'.format(exc))
            break

    client.delete_domain(generated_domain)

    completed += 2

    client.logout()

    return completed

def main():
    s = time.perf_counter()

    n_processes = 10
    with Pool() as pool:
        total = pool.map(run_test, range(n_processes))

    total = sum(total)

    elapsed = time.perf_counter() - s
    queries_per_sec = total/elapsed
    print(f"{__file__} executed in {elapsed:0.2f} seconds.")
    print(f"total queries = {total}")
    print(f"queries_per second = {queries_per_sec}")

if __name__ == "__main__":
    main()
