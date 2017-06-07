package com.tracker.client.activities;

import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;

import com.tracker.client.activities.widgets.FormCheckBox;
import com.tracker.client.activities.widgets.FormInputText;
import com.tracker.client.activities.widgets.FormSelect;
import com.tracker.client.controls.Component;
import com.tracker.client.jso.EditFormResp;
import com.tracker.client.jso.EditView;
import com.tracker.client.jso.Widget;
import com.tracker.client.util.DomUtils;
import com.tracker.client.util.JsArrayUtils;
import com.tracker.client.util.StringUtils;
import com.google.gwt.i18n.client.DateTimeFormat;

import elemental.dom.Element;
import elemental.json.JsonObject;
import elemental.json.JsonValue;

public class EditViewComponent extends Component {
  private static final Logger log = Logger.getLogger(EditViewComponent.class.getName());
  private static DateTimeFormat dateTimeFormat = DateTimeFormat.getFormat("yyyy-MM-dd'T'HH:mm");
  private static DateTimeFormat dateFormat = DateTimeFormat.getFormat("yyyy-MM-dd");
  private static DateTimeFormat timeFormat = DateTimeFormat.getFormat("HH:mm");
  private final EditView editView;
  private final EditFormResp editFormResp;
  private final Map<Widget, Object> mapWidgets = new HashMap<>();
  private JsonObject model;

  public EditViewComponent(EditFormResp editFormResp) {
    this.editFormResp = editFormResp;
    this.editView = editFormResp.getEditView();
  }

  @Override
  protected void createDom() {
    decorateInternal(DomUtils.createDom(doc.createDivElement()));
  }

  public void updateView(JsonObject model) {
    this.model = model;
    for (Map.Entry<Widget, Object> entry : mapWidgets.entrySet()) {
      final Widget description = entry.getKey();
      String id = description.getId();
      final Object widget = entry.getValue();
      JsonValue value = model.get(id);
      if (widget instanceof FormSelect) {
        ((FormSelect)widget).setSelectedValue(value.asString());
      } else if (widget instanceof FormCheckBox) {
        ((FormCheckBox)widget).setChecked(value.asBoolean());
      } else if (widget instanceof FormInputText) {
        switch (description.getType()) {
          case "TEXT_BOX_NUMBER": {
            ((FormInputText)widget).setValueDouble(value.asNumber());
            break;
          }
          case "TEXT_BOX_DATETIME": {
            ((FormInputText)widget).setValue(setValueDateTime(value.asString(), description.getDefaultValue()));
            break;
          }
          case "TEXT_BOX_DATE": {
            ((FormInputText)widget).setValue(setValueDate(value.asString(), description.getDefaultValue()));
            break;
          }
          case "TEXT_BOX_TIME": {
            ((FormInputText)widget).setValue(setValueTime(value.asString(), description.getDefaultValue()));
            break;
          }
          default: {
            ((FormInputText)widget).setValue(value.asString());
          }
        }
      }
    }
  }

  private String setValueDateTime(String val, String defaultValue) {
    val = goDateToJavaDate(val);
    if (val.length() == 0) {
      if ("now()".equalsIgnoreCase(defaultValue)) {
        return dateTimeFormat.format(new Date());
      }
    }
    return val;
  }

  private String setValueDate(String val, String defaultValue) {
    val = goDateToJavaDate(val);
    if (val.length() == 0) {
      if ("now()".equalsIgnoreCase(defaultValue)) {
        return dateFormat.format(new Date());
      }
    }
    return val.substring(0, 10);
  }

  private String setValueTime(String val, String defaultValue) {
    if (val == null || val.length() == 0) {
      if ("now()".equalsIgnoreCase(defaultValue)) {
        return timeFormat.format(new Date());
      }
    }
    if (val.length() < 4) {
      val = "0000".substring(val.length()) + val;
    }
    return val.substring(0, 2)  + ":" + val.substring(2);
  }

  private String getValueTime(String val) {
    return val.substring(0, 2) + val.substring(3);
  }

  private String goDateToJavaDate(String str) {
    return str != null ? str.substring(0, 16) : ""; 
  }

  private String getValueDateTime(String value) {
    if (StringUtils.hasLength(value)) {
      return javaDateToGoDate(value);
    }
    return "";
  }

  private String getValueDate(String value) {
    if (StringUtils.hasLength(value)) {
      return javaDateToGoDate(value + "T00:00");
    }
    return "";
  }

  private String javaDateToGoDate(String str) {
    return str + ":00Z";
  }

  public JsonObject updateModel() {
    if (model == null) {
      return null;
    }
    for (Map.Entry<Widget, Object> entry : mapWidgets.entrySet()) {
      final Widget description = entry.getKey();
      String id = description.getId();
      final Object widget = entry.getValue();

      if (widget instanceof FormSelect) {
        model.put(id, ((FormSelect)widget).getSelectedValue());
      } else if (widget instanceof FormCheckBox) {
        model.put(id, ((FormCheckBox)widget).getChecked());
      } else if (widget instanceof FormInputText) {
        switch (description.getType()) {
          case "TEXT_BOX_NUMBER": {
            model.put(id, ((FormInputText)widget).getValueDouble());
            break;
          }
          case "TEXT_BOX_DATETIME": {
            model.put(id, getValueDateTime(((FormInputText)widget).getValue()));
            break;
          }
          case "TEXT_BOX_DATE": {
            model.put(id, getValueDate(((FormInputText)widget).getValue()));
            break;
          }
          case "TEXT_BOX_TIME": {
            model.put(id, getValueTime(((FormInputText)widget).getValue()));
            break;
          }
          default: {
            model.put(id, ((FormInputText)widget).getValue());
          }
        }
      } else {
        log.info("unknown widget for field " + id);
      }
    }
    String v = model.toJson();
    log.info("updateModel = " + v);
    return model;
  }

  @Override
  public void decorateInternal(final Element element) {
    super.decorateInternal(element);

    for (Widget widget : JsArrayUtils.toList(editView.getWidgets())) {
      // log.info("widget = " + widget.getType());
      switch (widget.getType()) {
        case "LIST_BOX": {
          FormSelect input = new FormSelect()
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setOptions(editFormResp.getReferences(widget.getId()));
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX_NUMBER": {
          FormInputText input = new FormInputText() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setType("text");
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX_CHECKBOX": {
          FormCheckBox input = new FormCheckBox() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setDefaultValue(widget.getDefaultValue());
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX_DATETIME": {
          FormInputText input = new FormInputText() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setPlaceholder("2009-12-24 23:59")
            .setType("datetime-local");
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX_DATE": {
          FormInputText input = new FormInputText() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setPlaceholder("2009-12-24")
            .setType("date");
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX_TIME": {
          FormInputText input = new FormInputText() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setPlaceholder("23:59")
            .setType("time");
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
        case "TEXT_BOX": {
          FormInputText input = new FormInputText() 
            .setName(widget.getId())
            .setTitle(widget.getTitle())
            .setType("text");
          mapWidgets.put(widget, input);
          addChild(input, true);
          break;
        }
      }
    }
  }
}
