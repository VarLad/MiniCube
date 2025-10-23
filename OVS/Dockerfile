FROM debian

RUN apt -y update && apt -y install openvswitch-switch iproute2
 
ENV PATH="$PATH:/usr/share/openvswitch/scripts"

CMD ["ovs-ctl", "start"]
