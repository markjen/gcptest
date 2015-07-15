tail:
	tail -f .var/output

clean:
	rm -rf .var

serve:
	mkdir -p .var
	dev_appserver.py --storage_path=.var default/dispatch.yaml default/app.yaml worker/app.yaml 2>&1 | tee -ai .var/output

deploy:
	appcfg.py update default  # Updates default module, queues, and dispatch
	appcfg.py update worker/app.yaml
