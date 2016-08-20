#!/bin/sh
find . -name "*.pyc" -exec rm '{}' ';'
find . -name "*.o" -exec rm '{}' ';'
find . -name "*.a" -exec rm '{}' ';'
find . -name "*.so" -exec rm '{}' ';'
