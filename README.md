<h2 align="center">Service for birthday notifications</h2>

<p align="center">
  <img src="https://github.com/ilchenkoss/rutube_/blob/main/example.gif" alt="Example">
</p>

##  

### make bin file with Makefile:
```sh
make server
```

### bot commands:

```text
/subscribeToNotifications "true" for turn on and "false" for turn off notifications
```
```text
/subscribeTo "telegram_id" or "@username" for subscribe to user birthday
```
```text
/unSubscribeFrom "telegram_id" or "@username" for unsubscribe from user
```

### Please enter:

.env
```text
API_TOKEN telegram api token
```

config.yaml
```text
birthday_group_id group to birthday telegram id
group_owner_id group owner telegram id
```