clean:
	rm -rf ../var/scaling

serve:
	mkdir -p ../var/scaling
	dev_appserver.py --storage_path=../var/scaling default/dispatch.yaml default/app.yaml worker/app.yaml
