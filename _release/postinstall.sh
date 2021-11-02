#!/bin/bash
systemctl daemon-reload
systemctl enable beacon.service
systemctl restart beacon.service
