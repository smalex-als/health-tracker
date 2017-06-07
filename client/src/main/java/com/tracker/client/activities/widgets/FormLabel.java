package com.tracker.client.activities.widgets;

import com.tracker.client.controls.Component;

public class FormLabel extends Component {

	public void setText(String text) {
    if (getElement() == null) {
      createDom();
    }
    getElement().setInnerHTML(text);
	}
}
