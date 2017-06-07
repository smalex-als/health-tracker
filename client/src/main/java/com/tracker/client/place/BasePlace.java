package com.tracker.client.place;

import java.util.HashMap;
import java.util.Map;

import com.tracker.client.AppFactory;
import com.tracker.client.util.NumberUtils;
import com.tracker.client.util.StringUtils;
import com.google.gwt.place.shared.Place;
import com.google.gwt.place.shared.PlaceTokenizer;
import com.google.gwt.place.shared.Prefix;
import java.util.logging.Logger;


public class BasePlace extends Place {
  @Prefix("l")
  public static class Tokenizer implements PlaceTokenizer<BasePlace> {
    private static final Logger log = Logger.getLogger(BasePlace.class.getName());
    private static final String SEPARATOR = "/";
    private static final String SEPARATOR_PARENT = "!";
    private final AppFactory factory;

    public Tokenizer(AppFactory factory) {
      this.factory = factory;
    }

    public BasePlace getPlace(String token) {
      log.info("getPlace = " + token);
      int commaIndex = token.indexOf(':');
      String namespaceToken = null;
      if (commaIndex > 0) {
        namespaceToken = token.substring(0, commaIndex);
        token = token.substring(commaIndex + 1);
      }
      return getPlace(namespaceToken, token);
    }

    private BasePlace getPlace(String namespaceToken, String token) {
      Builder builder = newBuilder();
      int i = token.indexOf(SEPARATOR_PARENT);
      if (i > 0) {
        String parentToken = token.substring(0, i);
        builder.parent(getPlace(namespaceToken, parentToken));
        token = token.substring(i + SEPARATOR_PARENT.length());
      }
      int indexParam = token.indexOf(SEPARATOR);
      String typeToken;
      String paramToken = null;
      if (indexParam > 0) {
        typeToken = token.substring(0, indexParam);
        paramToken = token.substring(indexParam + SEPARATOR.length());
      } else {
        typeToken = token;
      }
      ViewId viewId = new ViewId(typeToken);
      builder.viewId(viewId);
      if (!typeToken.contains("edit")) {
        builder.params(StringUtils.paramsFromString(paramToken));
      } else {
        if (paramToken != null) {
          int indexParam2 = paramToken.indexOf(SEPARATOR);
          if (indexParam2 > 0) {
            builder.id(paramToken.substring(0, indexParam2));
            builder.rev(paramToken.substring(indexParam2 + SEPARATOR.length()));
          } else {
            builder.id(paramToken);
          }
        }
      }
      return builder.build();
    }

    @Override
    public String getToken(BasePlace place) {
      StringBuilder sb = new StringBuilder();
      if (place.getParent() != null) {
        sb.append(getToken(place.getParent()));
        sb.append(SEPARATOR_PARENT);
      }
      sb.append(place.getViewId().getActivityId());
      if (place.id != null && place.id.length() > 0) {
        sb.append(SEPARATOR);
        sb.append(place.id);
        if (place.rev != null && place.rev.length() > 0) {
          sb.append(SEPARATOR).append(place.rev);
        }
      } else if (place.params != null && !place.params.isEmpty()) {
        sb.append(SEPARATOR);
        sb.append(StringUtils.paramsToString(place.params));
      }
      return sb.toString();
    }
  }

  public static class Builder {
    private BasePlace parent;
    private ViewId viewId;
    private String id = "";
    private String rev = "";
    private Map<String, String> params = new HashMap<String, String>();

    public Builder parent(BasePlace parent) {
      this.parent = parent;
      return this;
    }

    public Builder viewId(ViewId viewId) {
      this.viewId = viewId;
      return this;
    }

    public Builder id(String id) {
      this.id = id;
      return this;
    }

    public Builder rev(String rev) {
      this.rev = rev;
      return this;
    }

    public Builder param(String name, String value) {
      if (StringUtils.hasText(value)) {
        params.put(name, value.trim());
      }
      return this;
    }

    public Builder paramInt(String name, int value) {
      params.put(name, String.valueOf(value));
      return this;
    }

    public Builder clearParam(String name) {
      params.remove(name);
      return this;
    }

    public Builder params(Map<String, String> params) {
      this.params = params;
      return this;
    }

    public BasePlace build() {
      return new BasePlace(viewId, id, rev, parent, params);
    }
  }

  private BasePlace(ViewId viewId, String id, String rev, BasePlace parent, Map<String, String> params) {
    this.viewId = viewId;
    this.id = id;
    this.rev = rev;
    this.parent = parent;
    this.params = params;
  }

  private String rev;
  private BasePlace parent;
  private ViewId viewId;
  private String id;
  private Map<String, String> params;

  public static Builder newBuilder() {
    return new Builder();
  }

  public ViewId getViewId() {
    return viewId;
  }

  public String getId() {
    return id;
  }

  public String getRev() {
    return rev;
  }

  public Map<String, String> getParams() {
    return params;
  }

  public BasePlace getParent() {
    return parent;
  }

  public int getParamInt(String name) {
    return NumberUtils.toInt(params.get(name));
  }

  public String getParam(String name) {
    return params.get(name);
  }

  @Override
  public String toString() {
    return "BasePlace{" +
        "parent=" + parent +
        ", viewId=" + viewId +
        ", id='" + id + '\'' +
        ", rev='" + rev + '\'' +
        ", params=" + params +
        '}';
  }
}
