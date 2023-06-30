import os
import sys
import mimetypes
from string import Template

vals = {
    'epp_ns':'http://www.ripn.net/epp/ripn-epp-1.0',
    'eppcom_ns':'http://www.ripn.net/epp/ripn-eppcom-1.0',
    'domain_ns':'http://www.ripn.net/epp/ripn-domain-1.0',
    'host_ns':'http://www.ripn.net/epp/ripn-host-1.0',
    'contact_ns':'http://www.ripn.net/epp/ripn-contact-1.0',
    'registrar_ns':'http://www.ripn.net/epp/ripn-registrar-1.0',
}

def copy_file(source, target):
    if mimetypes.guess_type(source)[0].split('/')[0] != "text":
        sys.stderr.write(f'not a text file {source}\n')
        return

    with open(source, 'r') as f:
        content = f.read()

    t = Template(content)
    content = t.substitute(**vals)

    with open(target, 'w') as f:
        f.write(content)

def main():
    source = sys.argv[1]
    target = sys.argv[2]

    mimetypes.init()
    mimetypes.add_type("text/xml", ".xsd")
    mimetypes.add_type("text/xml", ".conf")

    if os.path.isfile(source):
        # if we are copying into directory, format target as a file
        if os.path.isdir(target):
            target = target.rstrip('/')
            filename = source.split('/')[-1]
            target += '/' + filename
        copy_file(source, target)

    elif os.path.isdir(source):
        if os.path.isfile(target):
            raise Exception(f'cannot copy directory into a file : {target}')
        source = source.rstrip('/')

        # create a new directory in target dir
        dirname = source.split('/')[-1]
        if not os.path.exists(f'{target}/{dirname}'):
            os.mkdir(f'{target}/{dirname}')
          
        for file in os.listdir(source):
            copy_file(f'{source}/{file}', f'{target}/{dirname}/{file}')

    else:
        raise Exception(f'source does not exist {source}')

if __name__ == '__main__':
    main()
