FROM debian

RUN apt -y update && apt -y install openvswitch-switch iproute2
 
CMD ["sh", "-c", "ovs-ctl start; bash"]

