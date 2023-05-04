import asyncio
import time
import random

class WhoisClient(asyncio.Protocol):
    def __init__(self, consumer):
        self.consumer = consumer

    def connection_made(self, transport):
        self.consumer.on_connection(transport)

    def data_received(self, data):
        self.consumer.on_receive(data)

    def connection_lost(self, exc):
        self.consumer.connection_lost()

class WhoisConsumer():
    def __init__(self, on_conn_lost, max_messages=10, domains=None):
        self.on_conn_lost = on_conn_lost
        self.n_messages = 0
        self.max_messages = max_messages
        self.transport = None
        self.domains = domains

    def set_transport(self, transport):
        self.transport = transport

    def close_transport(self):
        if self.transport is not None:
            self.transport.close()

    def on_receive(self, data):
#        print(self.consumer_n) #data.decode())
        self.n_messages += 1
        if self.n_messages >= self.max_messages:
            self.close_transport()
            return
        self.send_message()

    def send_message(self):
        if self.transport is None:
            return
        selected_domain = random.choice(self.domains)
        message = '{}\n'.format(selected_domain)
        self.transport.write(message.encode())

    def on_connection(self, transport):
        self.set_transport(transport)
        self.send_message()

    def connection_lost(self):
        self.close_transport()
        self.on_conn_lost.set_result(self.n_messages)

async def main():
    messages = 1000
    tasks = 20
    domains = ['new-domain5.net.ru', 'new-delegated.net.ru', 'nonexistant.net.ru']

    loop = asyncio.get_running_loop()
    futures = []

    s = time.perf_counter()

    for i in range(tasks):
        on_conn_lost = loop.create_future()
        
        consumer = WhoisConsumer(on_conn_lost, max_messages=messages, domains=domains)

        transport, protocol = await loop.create_connection(
                    lambda: WhoisClient(consumer),
                    '127.0.0.1', 7072)
        futures.append(on_conn_lost)

    results = await asyncio.gather(*futures)
    total = sum(results)

    elapsed = time.perf_counter() - s
    queries_per_sec = total/elapsed

    print(f"total queries = {total}")
    print(f"queries_per second = {queries_per_sec}")

if __name__ == '__main__':
    asyncio.run(main())
