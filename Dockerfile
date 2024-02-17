FROM python:3.11

WORKDIR /code

# Copy only the requirements.txt first to leverage Docker cache
COPY ./requirements.txt /code/

RUN pip install --no-cache-dir --upgrade -r /code/requirements.txt

# Copy your app code into the WORKDIR
COPY . /code/

CMD ["uwsgi", "--ini", "/code/uwsgi.ini", "--http-socket=0.0.0.0:80"]

