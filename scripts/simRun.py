#!/usr/bin/python
#
# Command line arguments:
# [1] -- Number of server nodes
# [2] -- Name of result file

import os
from sys import argv
from hostnames import hostip, hostmach
import socket

dashboard = None
home_directory = "/home/ubuntu"
PATH = os.getcwd()
#result_dir = PATH + "/results/"
result_dir = home_directory+"/resilientdb/results/"
verifier_ip="129.159.37.193"

# Total nodes
nds = int(argv[1])
resfile = argv[2]
run = int(argv[3])
send_ifconfig = True

#cmd = "mkdir -p {}".format(result_dir)
# os.system(cmd)
# os.system("cp config.h {}".format(result_dir))
os.system("cp config.h {}".format(result_dir + "config_" + resfile))


machines = hostip
mach = hostmach

#	# check all rundb/runcl are killed
cmd = './vcloud_cmd.sh \"{}\" \"pkill -f \'rundb\'\"'.format(' '.join(machines))
print(cmd)
os.system(cmd)
cmd = './vcloud_cmd.sh \"{}\" \"pkill -f \'runcl\'\"'.format(' '.join(machines))
print(cmd)
os.system(cmd)
cmd = './vcloud_cmd.sh \"{}\" \"pkill -f \'invoker\'\"'.format(' '.join(machines))
print(cmd)
os.system(cmd)
cmd = './vcloud_cmd.sh \"{}\" \"pkill -f \'v_recv\'\"'.format(' '.join(machines))
print(cmd)
os.system(cmd)
cmd = './vcloud_cmd.sh \"{}\" \"sudo pkill -f \'verifier\'\"'.format(verifier_ip)
print(cmd)
os.system(cmd)

# cmd = './vcloud_cmd.sh \"{}\" \"mkdir -p \'resilientdb\'\"'.format(' '.join(machines))
# print(cmd)

# if run == 0:
os.system("./scp_binaries.sh {} {} {}".format(nds, 1 if send_ifconfig else 0, verifier_ip))

# '''

# if dashboard is not None:
#     hostname = socket.gethostname()
#     IPAddr = socket.gethostbyname(hostname)
#     print("Influx IP Address is:", IPAddr)

#     # check if monitorResults.sh processes are killed
#     os.system(
#         './vcloud_cmd.sh \"{}\" \"pkill -f \'sh monitorResults.sh\'\"'.format(' '.join(machines)))

#     # running monitorResults
#     cmd_monitor = './vcloud_monitor.sh \"{}\" \"{}\"'.format(
#         ' '.join(machines), IPAddr)
#     print(cmd_monitor)
#     os.system(cmd_monitor)

# # running the experiment
cmd = './vcloud_deploy.sh \"{}\" {} \"{}\" {}'.format(' '.join(machines), nds, resfile, verifier_ip)
print(cmd)
os.system(cmd)

# if dashboard is not None:
#     # check if monitorResults.sh processes are killed
#     os.system(
#         './vcloud_cmd.sh \"{}\" \"pkill -f \'sh monitorResults.sh\'\"'.format(' '.join(machines)))

# # collecting the output
os.system("./scp_results.sh {} {} {} {}".format(nds, resfile, result_dir, verifier_ip))


# '''